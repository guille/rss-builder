// Package rss provides types and functions for generating RSS 2.0 XML feeds.
package rss

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const PubDateFormat = "Mon, 02 Jan 2006 15:04:05 MST" // RFC-822 with 4-digit year

const contentNS = "http://purl.org/rss/1.0/modules/content/"

type RSS struct {
	XMLName   xml.Name `xml:"rss"`
	Version   string   `xml:"version,attr"`
	ContentNS string   `xml:"xmlns:content,attr"`
	Channel   Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Content     *CDATA `xml:"content:encoded"`
}

// CDATA holds a value that must not be entity-escaped, such as an article body
type CDATA struct {
	Value string `xml:",cdata"`
}

// NewCDATA returns nil for blank input so the element is left out altogether
func NewCDATA(s string) *CDATA {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &CDATA{Value: s}
}

func NewGUID(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func Write(w io.Writer, channel Channel) error {
	feed := RSS{
		Version:   "2.0",
		ContentNS: contentNS,
		Channel:   channel,
	}
	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rss: %w", err)
	}
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	_, err = w.Write(output)
	return err
}
