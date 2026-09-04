package sites

import (
	"testing"
	"time"
)

func TestParseUpgradeURL(t *testing.T) {
	const base = "https://ereaderfiles.kobo.com/firmwares/kobo9/"

	tests := []struct {
		name        string
		url         string
		wantVersion string
		wantDate    time.Time
		wantErr     bool
	}{
		{
			name:        "current",
			url:         base + "May2026/kobo-update-4.38.23697.zip",
			wantVersion: "4.38.23697",
			wantDate:    time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "build suffix",
			url:         base + "Oct2021/kobo-update-4.29.18730-TF6408.zip",
			wantVersion: "4.29.18730",
			wantDate:    time.Date(2021, time.October, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "codename instead of month",
			url:     "https://ereaderfiles.kobo.com/firmwares/kobo3/cod/kobo3-update-1.9.17.zip",
			wantErr: true,
		},
		{
			name:    "no version",
			url:     base + "May2026/kobo-update.zip",
			wantErr: true,
		},
	}

	var p KoboParser
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, released, err := p.parseUpgradeURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("got version %q, want an error", version)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseUpgradeURL: %v", err)
			}
			if version != tt.wantVersion {
				t.Errorf("got version %q, want %q", version, tt.wantVersion)
			}
			if !released.Equal(tt.wantDate) {
				t.Errorf("got date %s, want %s", released, tt.wantDate)
			}
		})
	}
}
