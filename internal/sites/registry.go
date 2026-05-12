package sites

import (
	"net/http"
)

func BuildAll(httpClient *http.Client) []Parser {
	return []Parser{
		GhosttyParser{httpClient: httpClient},
		AlbiacParser{httpClient: httpClient},
		SutherlandParser{httpClient: httpClient},
		KirshatrovParser{httpClient: httpClient},
	}
}
