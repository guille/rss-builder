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
		SanmiguelParser{httpClient: httpClient},
		NewCriterionDispatchParser{httpClient: httpClient},
		KoboParser{httpClient: httpClient, device: koboLibra2Device, label: "Kobo Libra 2"},
	}
}
