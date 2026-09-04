package sites

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	"go.guillerg.dev/rss-builder/internal/rss"
)

const (
	koboAPI      = "https://api.kobobooks.com/1.0"
	koboStuffURL = "https://pgaskin.net/KoboStuff/kobofirmware.html"
	// https://github.com/pgaskin/KoboStuff/blob/gh-pages/kfw.js
	koboLibra2Device = "00000000-0000-0000-0000-000000000388"

	// Kobo only answers what upgrade follows a given version, so the current
	// release has to be walked to from below the device's launch firmware. Any
	// version that old yields the same chain.
	koboSeedVersion = "1.0.0"

	koboAffiliate = "kobo"
	koboSerial    = "N0"
	koboMaxHops   = 12
)

// Tolerates the build suffix a few older releases carry, as in
// kobo-update-4.29.18730-TF6408.zip
var koboVersionRe = regexp.MustCompile(`-(?:update|upgrade)-(\d+\.\d+\.\d+)`)

type KoboParser struct {
	httpClient *http.Client
	device     string
	label      string
}

func (p KoboParser) Name() string     { return p.label + " firmware" }
func (KoboParser) URL() string        { return koboStuffURL }
func (KoboParser) dateFormat() string { return "Jan2006" }

type koboUpgradeCheck struct {
	UpgradeType    int    `json:"UpgradeType"`
	UpgradeURL     string `json:"UpgradeURL"`
	ReleaseNoteURL string `json:"ReleaseNoteURL"`
}

func (p KoboParser) Fetch() ([]rss.Item, error) {
	release, err := p.currentRelease()
	if err != nil {
		return nil, err
	}

	version, released, err := p.parseUpgradeURL(release.UpgradeURL)
	if err != nil {
		return nil, err
	}

	var content string
	if release.ReleaseNoteURL != "" {
		notes, nerr := p.releaseNotes(release.ReleaseNoteURL)
		if nerr != nil {
			log.Printf("%s: release notes for %s: %v", p.Name(), version, nerr)
		}
		content = notes
	}

	return []rss.Item{{
		Title:       p.label + " " + version,
		Link:        release.UpgradeURL,
		Description: "",
		GUID:        rss.NewGUID(release.UpgradeURL),
		PubDate:     released.Format(rss.PubDateFormat),
		Content:     rss.NewCDATA(content),
	}}, nil
}

// currentRelease walks the upgrade chain from koboSeedVersion and returns the
// last hop before Kobo stops offering one
func (p KoboParser) currentRelease() (koboUpgradeCheck, error) {
	version := koboSeedVersion
	seen := map[string]bool{version: true}

	var latest koboUpgradeCheck
	for range koboMaxHops {
		next, err := p.upgradeCheck(version)
		if err != nil {
			return koboUpgradeCheck{}, err
		}

		// UpgradeType only tells mandatory from optional, so a null URL is what
		// marks the device as up to date
		if next.UpgradeURL == "" {
			if latest.UpgradeURL == "" {
				return koboUpgradeCheck{}, fmt.Errorf("no upgrade offered for %s from %s", p.device, version)
			}
			return latest, nil
		}

		nextVersion, _, err := p.parseUpgradeURL(next.UpgradeURL)
		if err != nil {
			return koboUpgradeCheck{}, err
		}
		if seen[nextVersion] {
			return koboUpgradeCheck{}, fmt.Errorf("upgrade chain revisits %s", nextVersion)
		}
		seen[nextVersion] = true

		latest, version = next, nextVersion
	}
	return koboUpgradeCheck{}, fmt.Errorf("upgrade chain longer than %d hops", koboMaxHops)
}

// upgradeCheck asks what upgrade Kobo offers a device running version
func (p KoboParser) upgradeCheck(version string) (koboUpgradeCheck, error) {
	endpoint := fmt.Sprintf("%s/UpgradeCheck/Device/%s/%s/%s/%s",
		koboAPI, p.device, koboAffiliate, version, koboSerial)

	var check koboUpgradeCheck
	if err := fetchJSON(p.httpClient, endpoint, &check); err != nil {
		return koboUpgradeCheck{}, fmt.Errorf("upgrade check from %s: %w", version, err)
	}
	return check, nil
}

// releaseNotes returns the notes Kobo shows on the device before updating
func (p KoboParser) releaseNotes(noteURL string) (string, error) {
	doc, err := fetchDocument(p.httpClient, noteURL)
	if err != nil {
		return "", fmt.Errorf("fetch document: %w", err)
	}
	return articleHTML(doc, "body", noteURL)
}

// parseUpgradeURL pulls the version and release month out of a download URL
// like .../firmwares/kobo9/May2026/kobo-update-4.38.23697.zip. That folder is
// the only date the API carries, so items land on the first of the month.
func (p KoboParser) parseUpgradeURL(raw string) (string, time.Time, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("parse upgrade url %q: %w", raw, err)
	}

	match := koboVersionRe.FindStringSubmatch(u.Path)
	if match == nil {
		return "", time.Time{}, fmt.Errorf("no version in upgrade url %q", raw)
	}

	released, err := time.Parse(p.dateFormat(), path.Base(path.Dir(u.Path)))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("parse release month in upgrade url %q: %w", raw, err)
	}

	return match[1], released, nil
}
