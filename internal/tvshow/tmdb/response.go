package tmdb

import (
	"fmt"
	"net/http"

	"connectrpc.com/connect/v2"
)

// Unwrap returns the JSON200 body from an oapi-codegen response, mapping request
// failures and unexpected responses to Unavailable and HTTP 404 to NotFound.
func Unwrap[T, R any, P interface {
	*R
	GetJSON200() *T
	StatusCode() int
}](response P, err error) (*T, error) {
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err.Error())
	}
	if response == nil {
		return nil, connect.NewError(connect.CodeUnavailable, "not found")
	}
	status := response.StatusCode()
	if status == http.StatusNotFound {
		return nil, connect.NewError(connect.CodeNotFound, "not found")
	}
	if status != http.StatusOK || response.GetJSON200() == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Sprintf("unexpected upstream response (status %d)", status))
	}
	return response.GetJSON200(), nil
}
