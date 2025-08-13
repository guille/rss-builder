package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/guille/rss-builder/rss"
	"github.com/guille/rss-builder/sites/rory_sutherland"
)

type Parser interface {
	Name() string
	URL() string
	Fetch() ([]rss.Item, error)
}

func main() {
	parsers := []Parser{
		rory_sutherland.Parser{},
	}

	const outputDir = "output"
	if err := os.Mkdir(outputDir, 0o755); err != nil && !os.IsExist(err) {
		log.Fatalf("error creating output dir: %v", err)
	}

	for _, parser := range parsers {
		items, err := parser.Fetch()
		if err != nil {
			log.Printf("error fetching %s: %v", parser.Name(), err)
			continue
		}

		filename := filepath.Join(outputDir, parser.Name()+".xml")
		f, err := os.Create(filename)
		if err != nil {
			log.Printf("error creating %s: %v", filename, err)
			continue
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
		}
	}
}
