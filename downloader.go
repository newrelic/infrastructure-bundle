package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/google/go-github/v35/github"
	"gopkg.in/yaml.v3"
)

// config is the in-memory representation of the bundle.yml file
type config struct {
	AgentVersion string        `yaml:"agentVersion"`
	Integrations []integration `yaml:"integrations"`

	integrationConfig `yaml:",inline"` // Default fields for integrations
}

type integration struct {
	Name              string            `yaml:"name"`
	Version           string            `yaml:"version"`
	integrationConfig `yaml:",inline"`  // Per-integration overrides
	Arch              string            `yaml:"-"` // Used for convenience evaluating the template
	ArchReplacements  map[string]string `yaml:"archReplacements"`

	Subpath string `yaml:"subpath"` // Extract to this subfolder, rather than the virtual root
}

type integrationConfig struct {
	URL        string   `yaml:"url"`
	StagingUrl string   `yaml:"stagingUrl"`
	Repo       string   `yaml:"repo"`
	Archs      []string `yaml:"archs"`

	urlTemplate  *template.Template // used to store the URL template
	repoTemplate *template.Template // used to store the URL template
}

func main() {
	bfname := flag.String("bundle", "bundle.yml", "path to bundle.yml")
	outdir := flag.String("outdir", "out", "path to output directory")
	workers := flag.Int("workers", 4, "number of download threads")
	agentonly := flag.Bool("agent-version", false, "print agent version and exit")
	staging := flag.Bool("staging", false, "use stagingUrl")
	overrideLatest := flag.Bool("latest", false, "ignore version and download latest from GitHub")
	flag.Parse()

	bundleFile, err := os.Open(*bfname)
	if err != nil {
		log.Fatal(err)
	}

	conf := config{}
	err = yaml.NewDecoder(bundleFile).Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	// Print agent version and exit
	if *agentonly {
		fmt.Print(conf.AgentVersion)
		return
	}

	// Validate and expand config
	if err := conf.expand(*staging, *overrideLatest); err != nil {
		log.Fatal(err)
	}

	// Scan all archs defined in the integration list and create subfolders for them
	// This is done separately from the concurrent download step to avoid concurrency issues
	if err := mkdirArchs(*outdir, conf.Integrations); err != nil {
		log.Fatal(err)
	}

	// Concurrently download and extract integrations in the yaml file
	ichan := make(chan *integration, len(conf.Integrations))
	errchan := make(chan error, len(conf.Integrations))
	wg := &sync.WaitGroup{}
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			for i := range ichan {
				errchan <- i.download(*outdir)
			}
			wg.Done()
		}()
	}

	// Send integration specs to the workers
	for i := range conf.Integrations {
		ichan <- &conf.Integrations[i]
	}
	close(ichan)
	wg.Wait()

	// Gather errors, if any
errloop:
	for {
		select {
		case err := <-errchan:
			if err != nil {
				log.Fatalf("error fetching integrations: %v", err)
			}
		default:
			log.Printf("Integrations downloaded")
			break errloop
		}
	}

	log.Printf("Preparing tree for install...")
	if err := prepareTree(*outdir); err != nil {
		log.Fatal(err)
	}

	log.Printf("All done, integrations installed to '%s'", *outdir)
}

// expand extends defaults to integrations and performs basic validation
func (conf *config) expand(useStaging, overrideLatest bool) error {
	if useStaging {
		conf.URL = conf.StagingUrl
	}

	urlTemplate, err := template.New("url").Parse(conf.URL)
	if err != nil {
		return fmt.Errorf("evaluating global URL template: %v", err)
	}

	conf.urlTemplate = urlTemplate

	repoTemplate, err := template.New("repo").Parse(conf.Repo)
	if err != nil {
		return fmt.Errorf("evaluating global URL template: %v", err)
	}

	conf.repoTemplate = repoTemplate

	for i := range conf.Integrations {
		integration := &conf.Integrations[i]

		if err := integration.expand(&conf.integrationConfig); err != nil {
			return fmt.Errorf("expanding config for %q: %w", integration.Name, err)
		}

		if !overrideLatest {
			continue
		}

		if err := integration.overrideVersion(useStaging); err != nil {
			return fmt.Errorf("overrding version for %q: %w", integration.Name, err)
		}
	}

	return nil
}

func (i *integration) expand(defaults *integrationConfig) error {
	if i.Name == "" {
		return fmt.Errorf("cannot process integration with an empty name")
	}

	if i.Version == "" {
		return fmt.Errorf("cannot download '%s' with an empty version", i.Name)
	}

	var err error

	urlTemplate := defaults.urlTemplate

	if i.URL != "" {
		if urlTemplate, err = template.New("url").Parse(i.URL); err != nil {
			return fmt.Errorf("building custom url template: %v", err)
		}
	}

	i.urlTemplate = urlTemplate

	repoTemplate := defaults.repoTemplate

	if i.Repo != "" {
		if repoTemplate, err = template.New("repo").Parse(i.Repo); err != nil {
			return fmt.Errorf("building custom repo template: %v", err)
		}
	}

	i.repoTemplate = repoTemplate

	if len(i.Archs) == 0 {
		i.Archs = defaults.Archs
	}

	return nil
}

// download Expands the URL template for each integration arch and extracts it to outdir.
func (i *integration) download(outdir string) error {
	for _, arch := range i.Archs {
		if replArch, hasReplacement := i.ArchReplacements[arch]; hasReplacement {
			i.Arch = replArch
		} else {
			i.Arch = arch
		}

		urlbuf := &bytes.Buffer{}
		err := i.urlTemplate.Execute(urlbuf, i)
		if err != nil {
			return fmt.Errorf("error evaluating template: %v", err)
		}
		url := urlbuf.String()

		log.Printf("Hitting %s", url)
		response, err := http.Get(url)
		if err != nil {
			return err
		}

		defer response.Body.Close()

		if response.StatusCode >= 300 {
			return fmt.Errorf("got status %d when fetching %s", response.StatusCode, url)
		}

		destination := path.Join(outdir, arch)
		if i.Subpath != "" {
			destination = path.Join(destination, i.Subpath)
		}
		iname := url[strings.LastIndex(url, "/"):]
		log.Printf("Downloading and extracting %s", iname)
		// Iterating over archive/tar is too long to write, going the hacky way...
		cmd := exec.Command("tar", "-xz")
		cmd.Dir = destination
		cmd.Stdin = response.Body
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running tar: %v", err)
		}
	}

	return nil
}

func (i *integration) overrideVersion(includePrereleases bool) error {
	repobuf := bytes.Buffer{}

	if err := i.repoTemplate.Execute(&repobuf, i); err != nil {
		return fmt.Errorf("could not evaluate repo template: %w", err)
	}

	orgRepo := strings.Split(repobuf.String(), "/")
	if len(orgRepo) != 2 {
		return fmt.Errorf("bad format for org/repo: %s", i.Repo)
	}

	log.Printf("getting latest version for %s...", i.Name)
	gh := github.NewClient(http.DefaultClient)
	releases, _, err := gh.Repositories.ListReleases(context.Background(), orgRepo[0], orgRepo[1], nil)
	if err != nil {
		return fmt.Errorf("could not get releases for %s: %w", i.Repo, err)
	}

	if !includePrereleases {
		stableReleases := make([]*github.RepositoryRelease, 0, len(releases))
		for _, r := range releases {
			if r.Prerelease != nil && !*r.Prerelease {
				stableReleases = append(stableReleases, r)
			}
		}
		releases = stableReleases
	}

	if len(releases) == 0 {
		return fmt.Errorf("repo %s does not have any release", i.Repo)
	}

	// Sort most recent first
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].CreatedAt.After(releases[j].CreatedAt.Time)
	})

	namePtr := releases[0].TagName
	if namePtr == nil {
		return fmt.Errorf("tagName for latest release of %s is nil", i.Repo)
	}

	log.Printf("Found %s %s...", i.Name, *namePtr)
	i.Version = *namePtr

	return nil
}

// prepareTree cleans up *.sample and windows-related files
func prepareTree(outdir string) error {
	return filepath.Walk(outdir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil // Continue
		}

		if strings.HasSuffix(info.Name(), ".sample") {
			return os.Remove(path)
		}

		for _, pattern := range []string{"-win-", "README", "CHANGELOG", "LICENSE"} {
			if strings.Contains(info.Name(), pattern) {
				return os.Remove(path)
			}
		}
		return nil
	})
}

// mkdirArchs scans all archs present in the integrations list and creates subfolders for them
func mkdirArchs(outdir string, integrations []integration) error {
	paths := map[string]struct{}{}
	for _, i := range integrations {
		for _, a := range i.Archs {
			paths[a] = struct{}{}
			if i.Subpath != "" {
				paths[path.Join(a, i.Subpath)] = struct{}{}
			}
		}
	}

	for arch := range paths {
		if err := os.MkdirAll(path.Join(outdir, arch), 0o755); err != nil {
			return fmt.Errorf("cannot create %s/%s: %v", outdir, arch, err)
		}
	}

	return nil
}
