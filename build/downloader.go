package main

import (
	"bytes"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

type config struct {
	URL          string        `yaml:"url"`
	Archs        []string      `yaml:"archs"`
	Integrations []integration `yaml:"integrations"`
}

type integration struct {
	Name    string   `yaml:"name"`
	Version string   `yaml:"version"`
	Archs   []string `yaml:"archs"`
	URL     string   `yaml:"url"`

	ArchReplacements map[string]string `yaml:"archReplacements"`

	Arch        string `yaml:"-"` // Used for convenience evaluating the template
	urlTemplate *template.Template
}

func main() {
	bfname := flag.String("bundle", "bundle.yml", "path to bundle.yml")
	outdir := flag.String("outdir", "out", "path to output directory")
	workers := flag.Int("workers", 4, "number of download threads")
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

	// Validate and expand config
	if err := expandConfig(&conf); err != nil {
		log.Fatal(err)
	}

	// Scan all archs defined in the integration list and create subfolders for them
	// This is done separately from the concurrent download step to avoid concurrency issues
	if err := mkdirArchs(*outdir, conf.Integrations); err != nil {
		log.Fatal(err)
	}

	// Concurrently fetch and extract integrations in the yaml file
	ichan := make(chan *integration, len(conf.Integrations))
	errchan := make(chan error, len(conf.Integrations))
	wg := &sync.WaitGroup{}
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			for i := range ichan {
				errchan <- fetch(i, *outdir)
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

// expandConfig extends defaults to integrations and performs basic validation
func expandConfig(conf *config) error {
	// Build template for default URL
	globalTemplate, err := template.New("url").Parse(conf.URL)
	if err != nil {
		return fmt.Errorf("error evaluating global URL template: %v", err)
	}

	for i := range conf.Integrations {
		integration := &conf.Integrations[i]

		if integration.Name == "" {
			return fmt.Errorf("cannot fetch integrations[%d] with an empty name", i)
		}
		if integration.Version == "" {
			return fmt.Errorf("cannot fetch '%s' with an empty version", integration.Name)
		}

		if integration.URL != "" {
			integration.urlTemplate, err = template.New("url").Parse(integration.URL)
			if err != nil {
				return fmt.Errorf("error building custom template: %v", err)
			}
		} else {
			integration.urlTemplate = globalTemplate
		}

		if len(integration.Archs) == 0 {
			integration.Archs = conf.Archs
		}
	}

	return nil
}

// fetch Expands the URL template for integrations and invokes downloadAndExtract
func fetch(i *integration, outdir string) error {
	for _, arch := range i.Archs {
		urlbuf := &bytes.Buffer{}

		if replArch, hasReplacement := i.ArchReplacements[arch]; hasReplacement {
			i.Arch = replArch
		} else {
			i.Arch = arch
		}

		err := i.urlTemplate.Execute(urlbuf, i)
		if err != nil {
			return fmt.Errorf("error evaluating template: %v", err)
		}

		err = downloadAndExtract(urlbuf.String(), path.Join(outdir, arch))
		if err != nil {
			return fmt.Errorf("error downloading integration: %v", err)
		}
	}

	return nil
}

// downloadAndExtract Hits the supplied URL and extracts the contents of the tar archive to the supplied directory
func downloadAndExtract(url string, outdir string) error {
	log.Printf("Hitting %s", url)
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return fmt.Errorf("got status %d when fetching %s", response.StatusCode, url)
	}

	ct := response.Header.Get("content-type")
	switch ct {
	case "application/x-tar":
		fallthrough
	case "application/octet-stream":
		break
	default:
		return fmt.Errorf("unexpected contenty type '%s' for %s", ct, url)
	}

	iname := url[strings.LastIndex(url, "/"):]
	log.Printf("Downloading and extracting %s", iname)
	// Iterating over archive/tar is too long to write, going the hacky way...
	cmd := exec.Command("tar", "-xz")
	cmd.Dir = outdir
	cmd.Stdin = response.Body
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running tar: %v", err)
	}

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

		if strings.Contains(info.Name(), "-win-") {
			return os.Remove(path)
		}

		return nil
	})
}

// mkdirArchs scans all archs present in the integrations list and creates subfolders for them
func mkdirArchs(outdir string, integrations []integration) error {
	archs := map[string]struct{}{}
	for _, i := range integrations {
		for _, a := range i.Archs {
			archs[a] = struct{}{}
		}
	}

	for arch := range archs {
		if err := os.MkdirAll(path.Join(outdir, arch), 0755); err != nil {
			return fmt.Errorf("cannot create %s/%s: %v", outdir, arch, err)
		}
	}

	return nil
}
