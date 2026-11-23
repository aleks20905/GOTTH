package handlers

import (
	"bytes"
	"context"
	"testing"

	"goth/internal/middleware"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundHandler(t *testing.T) {
	assert := assert.New(t)

	handler := NewNotFoundHandler()

	req, err := http.NewRequest("GET", "/nope", nil)
	assert.NoError(err)

	nonces := middleware.Nonces{
		Htmx:            "nonce-1234",
		ResponseTargets: "nonce-5678",
		Tw:              "nonce-9101",
		HtmxCSSHash:     "sha256-pgn1TCGZX6O77zDvy0oTODMOxemn0oj0LeCnQTRj7Kg=",
	}
	req = req.WithContext(context.WithValue(req.Context(), middleware.NonceKey, nonces))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(http.StatusNotFound, rr.Code, "handler returned wrong status code")
	assert.True(bytes.Contains(rr.Body.Bytes(), []byte("Page not found")), "handler returned unexpected body")
}
