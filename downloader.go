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
	"time"

	"github.com/google/go-github/v38/github"
	"golang.org/x/oauth2"
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
	oldVersion        string            // Will be set to the old version if a new one is found
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
	overrideLatest := flag.Bool("override-latest", false, "ignore version and download latest from GitHub")
	checkLatest := flag.Bool("check-latest", false, "check for new versions and exit")
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
	if err := conf.expand(*staging, *overrideLatest || *checkLatest); err != nil {
		log.Fatal(err)
	}

	// Print new versions and exit.
	if *checkLatest {
		conf.printUpdates()

		return
	}

	// Scan all archs defined in the integration list and create subfolders for them
	// This is done separately from the concurrent download step to avoid concurrency issues
	if err := mkdirArchs(*outdir, conf.Integrations); err != nil {
		log.Fatal(err)
	}

	// Concurrently download and extract integrations in the yaml file
	ichan := make(chan *integration, len(conf.Integrations))
	wg := &sync.WaitGroup{}
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			for i := range ichan {
				err := i.download(*outdir)
				if err != nil {
					log.Fatalf("Error downloading integration: %v", err)
				}
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

	log.Printf("Preparing tree for install...")
	if err := prepareTree(*outdir); err != nil {
		log.Fatal(err)
	}

	log.Printf("All done, integrations installed to '%s'", *outdir)
}

// expand compiles templates, extends defaults to integrations and performs basic validation
func (conf *config) expand(useStaging, overrideLatest bool) error {
	if useStaging {
		conf.URL = conf.StagingUrl
	}

	if conf.URL == "" {
		return fmt.Errorf("global download URL template is empty")
	}

	if conf.Repo == "" {
		return fmt.Errorf("global repo name template is empty")
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

	// Build GithubClient and fetch releases
	// oauthClientFromEnv will return an authenticated client if `$GITHUB_TOKEN` is present, or the default otherwise
	gh := github.NewClient(oauthClientFromEnv())

	// Iterate over integrations expanding their configs as well
	for i := range conf.Integrations {
		integration := &conf.Integrations[i]

		if err := integration.expand(&conf.integrationConfig); err != nil {
			return fmt.Errorf("expanding config for %q: %w", integration.Name, err)
		}

		// Skip version override if flag is not present
		if !overrideLatest {
			continue
		}

		// Fetch latest version from GitHub and override the one present in the original config.
		if err := integration.overrideVersion(gh, useStaging); err != nil {
			return fmt.Errorf("overrding version for %q: %w", integration.Name, err)
		}
	}

	return nil
}

// printUpdates prints the integrations that have an update available.
func (conf *config) printUpdates() {
	for _, i := range conf.Integrations {
		if i.oldVersion == "" {
			continue
		}

		fmt.Printf("  - name: %s\n    version: %s\n", i.Name, i.Version)
	}
}

// expand performs validation and fills empty values with those defined in the integration config.
func (i *integration) expand(defaults *integrationConfig) error {
	if i.Name == "" {
		return fmt.Errorf("cannot process integration with an empty name")
	}

	var err error

	urlTemplate := defaults.urlTemplate

	// Build URL template if overridden
	if i.URL != "" {
		if urlTemplate, err = template.New("url").Parse(i.URL); err != nil {
			return fmt.Errorf("building custom url template: %v", err)
		}
	}

	i.urlTemplate = urlTemplate

	repoTemplate := defaults.repoTemplate

	// Build repo template if overridden
	if i.Repo != "" {
		if repoTemplate, err = template.New("repo").Parse(i.Repo); err != nil {
			return fmt.Errorf("building custom repo template: %v", err)
		}
	}

	i.repoTemplate = repoTemplate

	// Copy global arch list if not defined
	if len(i.Archs) == 0 {
		i.Archs = defaults.Archs
	}

	return nil
}

// download expands the URL template for each integration arch and extracts it to outdir
func (i *integration) download(outdir string) error {
	// Check for empty version here rather than when expanding config since version may also come from Github
	if i.Version == "" {
		return fmt.Errorf("cannot download '%s' with an empty version", i.Name)
	}

	// Different archs for the same integration are processed sequentially
	for _, arch := range i.Archs {
		// Process arch replacements in URL
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

		if !strings.HasPrefix(url, "https://") {
			return fmt.Errorf("refusing to download using insecure non-https url: %s", url)
		}

		log.Printf("Downloading %s", url)
		response, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("downloading %s (%s): %w", i.Name, arch, err)
		}

		defer response.Body.Close()

		if response.StatusCode >= 300 {
			return fmt.Errorf("got status %d when fetching %s", response.StatusCode, url)
		}

		// Prepare path to extract, outdir/$arch
		destination := path.Join(outdir, arch)
		// Append subpath if defined. Usually not required since tarballs are structured to be extracted in /.
		if i.Subpath != "" {
			destination = path.Join(destination, i.Subpath)
		}

		log.Printf("Downloading and extracting %s (%s)", i.Name, arch)
		// Invoke tar externally with pipe (simplifies code).
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

// overrideVersion fetches the tag name of the latest release (or prerelease) from Github
func (i *integration) overrideVersion(gh *github.Client, includePrereleases bool) error {
	// Evaluate repo template
	repobuf := &bytes.Buffer{}

	if err := i.repoTemplate.Execute(repobuf, i); err != nil {
		return fmt.Errorf("could not evaluate repo template: %w", err)
	}

	// Split in user/repo
	orgRepo := strings.Split(repobuf.String(), "/")
	if len(orgRepo) != 2 {
		return fmt.Errorf("bad format for org/repo: %s", i.Repo)
	}

	log.Printf("Fetching releases for %s...", i.Name)

	allReleases := make([]*github.RepositoryRelease, 0, 30) // GH returns max 30 releases per page
	for page := 1; page != 0; {
		releases, response, err := gh.Repositories.ListReleases(context.Background(), orgRepo[0], orgRepo[1], &github.ListOptions{
			Page: page,
		})
		if err != nil {
			return fmt.Errorf("could not get releases for %s: %w", i.Repo, err)
		}

		allReleases = append(allReleases, releases...)
		page = response.NextPage
	}

	releases := make([]*github.RepositoryRelease, 0, len(allReleases))
	for _, r := range allReleases {
		// Filter out pre-releases if `includePrereleases` is not set.
		if !includePrereleases && r.GetPrerelease() {
			log.Printf("skipping pre-release %s %s", i.Name, r.GetTagName())
			continue
		}

		// Filter releases published less than one hour ago, since it is likely that their pipeline is still running
		// and packages are not in the staging repo yet.
		age := time.Since(r.GetPublishedAt().Time)
		if age < 1*time.Hour {
			log.Printf("skipping %s %s as it's too young (%v)", i.Name, r.GetTagName(), age)
			continue
		}

		releases = append(releases, r)
	}

	if len(releases) == 0 {
		return fmt.Errorf("repo %s does not have any acceptable release", i.Repo)
	}

	// Sort most recent first
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].GetPublishedAt().After(releases[j].GetPublishedAt().Time)
	})

	newVersion := strings.TrimPrefix(releases[0].GetTagName(), "v")
	if newVersion == "" {
		return fmt.Errorf("tagName for latest release of %s is empty", i.Repo)
	}

	if i.Version != newVersion {
		log.Printf("%s %s -> %s", i.Name, i.Version, newVersion)
		i.oldVersion = i.Version
		i.Version = newVersion
	}

	return nil
}

// oauthClientFromEnv returns an OAuth client using the GITHUB_TOKEN env var if it's present, or http.DefaultClient otherwise
func oauthClientFromEnv() *http.Client {
	ghtoken := os.Getenv("GITHUB_TOKEN")
	if ghtoken == "" {
		return http.DefaultClient
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghtoken},
	)
	return oauth2.NewClient(context.Background(), ts)
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
	// Collect all archs defined in all integrations
	paths := map[string]struct{}{}
	for _, i := range integrations {
		for _, arch := range i.Archs {
			paths[arch] = struct{}{}
			// If a subpath is defined, we need to create it as well
			if i.Subpath != "" {
				paths[path.Join(arch, i.Subpath)] = struct{}{}
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
