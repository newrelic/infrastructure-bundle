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
	Integrations []integration `yaml:"integrations"`
}

type integration struct {
	Name    string   `yaml:"name"`
	Version string   `yaml:"version"`
	Archs   []string `yaml:"arch"`
	URL     string   `yaml:"url"`

	Arch string `yaml:"-"` // Internal only

	ArchReplacements map[string]string `yaml:"archReplacements"`
}

var defaultArchs = []string{"amd64", "arm64", "arm"}

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

	// Build template for default URL
	urlTemplate, err := template.New("url").Parse(conf.URL)
	if err != nil {
		log.Fatal(err)
	}

	// Scan all archs defined in the integration list and create subfolders for them
	// This is done separately from the concurrent download step to avoid concurrency issues
	if err := mkdirArchs(*outdir, conf.Integrations); err != nil {
		log.Fatal(err)
	}

	// Concurrently fetch and extract integrations in the yaml file
	ichan := make(chan *integration, len(conf.Integrations))
	wg := &sync.WaitGroup{}
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			for i := range ichan {
				err := fetch(i, urlTemplate, *outdir)
				if err != nil {
					log.Printf("error fetching integration: %v")
				}
			}
			wg.Done()
		}()
	}

	for i := range conf.Integrations {
		ichan <- &conf.Integrations[i]
	}
	close(ichan)
	wg.Wait()
	log.Printf("Integrations downloaded")

	log.Printf("Preparing tree for install")
	if err := prepareTree(*outdir); err != nil {
		log.Fatal(err)
	}
	log.Printf("All done, integrations installed to '%s'", *outdir)
}

// fetch Expands the URL template for integrations and invokes downloadAndExtract
func fetch(i *integration, urltmpl *template.Template, outdir string) error {
	if i.Name == "" {
		return fmt.Errorf("cannot fetch integration with an empty name")
	}
	if i.Version == "" {
		return fmt.Errorf("cannot fetch '%s' with an empty version", i.Name)
	}

	if i.URL != "" {
		overrideTmpl, err := template.New("url").Parse(i.URL)
		if err != nil {
			return fmt.Errorf("error building custom template: %v", err)
		}
		urltmpl = overrideTmpl
	}

	if len(i.Archs) == 0 {
		i.Archs = defaultArchs
	}

	for _, arch := range i.Archs {
		urlbuf := &bytes.Buffer{}

		i.Arch = arch

		// Apply arch replacement
		if newArch, found := i.ArchReplacements[i.Arch]; found {
			i.Arch = newArch
		}

		err := urltmpl.Execute(urlbuf, i)
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

	if ct := response.Header.Get("content-type"); ct != "application/x-tar" {
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
