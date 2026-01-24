package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/guille/rss-builder/rss"
	"github.com/guille/rss-builder/sites/gabriel_albiac"
	"github.com/guille/rss-builder/sites/kirshatrov"
	"github.com/guille/rss-builder/sites/rory_sutherland"
)

type Parser interface {
	Name() string
	URL() string
	Fetch() ([]rss.Item, error)
}

func main() {
	parsers := []Parser{
		gabriel_albiac.Parser{},
		rory_sutherland.Parser{},
		kirshatrov.Parser{},
	}

	const outputDir = "output"
	if err := os.Mkdir(outputDir, 0o755); err != nil && !os.IsExist(err) {
		log.Fatalf("error creating output dir: %v", err)
	}

	errCh := make(chan error, len(parsers))

	for _, parser := range parsers {
		go func() {
			items, err := parser.Fetch()
			if err != nil {
				log.Printf("error fetching %s: %v", parser.Name(), err)
				errCh <- err
				return
			}

			filename := filepath.Join(outputDir, parser.Name()+".xml")
			f, err := os.Create(filename)
			if err != nil {
				log.Printf("error creating %s: %v", filename, err)
				errCh <- err
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
				log.Printf("error writing %s: %v", filename, err)
				errCh <- err
				return
			}

			errCh <- nil
		}()
	}

	var anyErr bool

	for range parsers {
		if err := <-errCh; err != nil {
			anyErr = true
		}
	}

	if anyErr {
		os.Exit(1)
	}
}
