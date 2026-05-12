// Package model contains shared models used by the program
package model

import (
	"github.com/guille/rss-builder/internal/rss"
)

type Parser interface {
	Name() string
	URL() string
	Fetch() ([]rss.Item, error)
}
