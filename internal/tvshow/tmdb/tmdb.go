package tmdb

import "net/http"

//go:generate go tool oapi-codegen -config cfg.yaml https://developer.themoviedb.org/openapi/tmdb-api.json

type AuthorizationTransport struct {
	http.RoundTripper
	bearer string
}

func (t *AuthorizationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.bearer)
	return t.RoundTripper.RoundTrip(req)
}

func NewTMDBClient(apiKey string) (*ClientWithResponses, error) {
	return NewClientWithResponses(ServerUrlHttpsapiThemoviedbOrg, WithHTTPClient(&http.Client{
		Transport: &AuthorizationTransport{
			RoundTripper: http.DefaultTransport,
			bearer:       apiKey,
		},
	}))
}

func PosterPath(posterPath string) string {
	return "https://image.tmdb.org/t/p/w500" + posterPath
}
