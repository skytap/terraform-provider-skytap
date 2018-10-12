package utils

import (
	"net/http"
	"testing"

	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/stretchr/testify/assert"
)

func TestResponseErrorIsNotFound_BadRequest(t *testing.T) {
	resp := skytap.ErrorResponse{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
		},
	}

	assert.False(t, ResponseErrorIsNotFound(&resp))
}

func TestResponseErrorIsNotFound_NotFound(t *testing.T) {
	resp := skytap.ErrorResponse{
		Response: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}

	assert.True(t, ResponseErrorIsNotFound(&resp))
}
