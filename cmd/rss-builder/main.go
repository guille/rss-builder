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

	"go.guillerg.dev/rss-builder/internal/rss"
	"go.guillerg.dev/rss-builder/internal/sites"
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
			errCh <- buildFeed(parser, outputDir)
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

// buildFeed scrapes parser and writes its feed into outputDir
func buildFeed(parser sites.Parser, outputDir string) (err error) {
	items, err := parser.Fetch()
	if err != nil {
		return fmt.Errorf("fetching %s: %v", parser.Name(), err)
	}

	filename := filepath.Join(outputDir, parser.Name()+".xml")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating %s: %v", filename, err)
	}
	// A write can fail at Close, and a truncated feed must not pass for success
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing %s: %v", filename, cerr)
		}
	}()

	channel := rss.Channel{
		Title:       parser.Name() + " feed",
		Link:        parser.URL(),
		Description: "Scraped feed for " + parser.Name(),
		Items:       items,
	}

	if err := rss.Write(f, channel); err != nil {
		return fmt.Errorf("writing %s: %v", filename, err)
	}
	return nil
}
