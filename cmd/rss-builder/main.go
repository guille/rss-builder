package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/guille/rss-builder/internal/rss"
	"github.com/guille/rss-builder/internal/sites"
)

func main() {
	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	parsers := sites.BuildAll(httpClient)

	const outputDir = "output"
	if err := os.Mkdir(outputDir, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
		log.Fatalf("error creating output dir: %v", err)
	}

	errCh := make(chan error, len(parsers))

	for _, parser := range parsers {
		go func() {
			items, err := parser.Fetch()
			if err != nil {
				errCh <- fmt.Errorf("fetching %s: %v", parser.Name(), err)
				return
			}

			filename := filepath.Join(outputDir, parser.Name()+".xml")
			f, err := os.Create(filename)
			if err != nil {
				errCh <- fmt.Errorf("creating %s: %v", filename, err)
				return
			}
			defer f.Close()

			channel := rss.Channel{
				Title:       parser.Name() + " feed",
				Link:        parser.URL(),
				Description: "Scraped feed for " + parser.Name(),
				Items:       items,
			}

			if err := rss.Write(f, channel); err != nil {
				errCh <- fmt.Errorf("writing %s: %v", filename, err)
				return
			}

			errCh <- nil
		}()
	}

	var anyErr bool

	for range parsers {
		if err := <-errCh; err != nil {
			log.Printf("error: %v", err)
			anyErr = true
		}
	}

	if anyErr {
		os.Exit(1)
	}
}
