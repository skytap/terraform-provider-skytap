package utils

import (
	"net/http"

	"github.com/skytap/skytap-sdk-go/skytap"
)

// ResponseErrorIsNotFound returns true if the error is NotFound
func ResponseErrorIsNotFound(responseError error) bool {
	if r, ok := responseError.(*skytap.ErrorResponse); ok {
		if r.Response.StatusCode == http.StatusNotFound {
			return true
		}
	}

	return false
}

// ResponseErrorIsUnprocessableEntity returns true if the error is UnprocessableEntity
func ResponseErrorIsUnprocessableEntity(responseError error) bool {
	if r, ok := responseError.(*skytap.ErrorResponse); ok {
		if r.Response.StatusCode == http.StatusUnprocessableEntity {
			return true
		}
	}

	return false
}
