package rss

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
)

const PubDateFormat = "Mon, 02 Jan 2006 15:04:05 MST" // RFC-822 with 4-digit year

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
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
}

func NewGUID(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func Write(w io.Writer, channel Channel) error {
	rss := RSS{
		Version: "2.0",
		Channel: channel,
	}
	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rss: %w", err)
	}
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	_, err = w.Write(output)
	return err
}
