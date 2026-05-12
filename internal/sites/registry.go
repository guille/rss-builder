package sites

import (
	"net/http"

	model "github.com/guille/rss-builder/internal/model"
)

func BuildAll(httpClient *http.Client) []model.Parser {
	return []model.Parser{
		GhosttyParser{httpClient: httpClient},
		AlbiacParser{httpClient: httpClient},
		SutherlandParser{httpClient: httpClient},
		KirshatrovParser{httpClient: httpClient},
	}
}
