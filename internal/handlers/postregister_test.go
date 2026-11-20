package handlers

import (
	"bytes"
	"context"

	"goth/internal/middleware"
	"goth/internal/store/dbstore"
	storemock "goth/internal/store/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterUserHandler(t *testing.T) {

	testCases := []struct {
		name               string
		email              string
		password           string
		createUserError    error
		expectedStatusCode int
		expectedBody       []byte
	}{
		{
			name:               "success",
			email:              "test@example.com",
			password:           "password",
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte("Account Created!"),
		},
		{
			name:               "fail - error creating user",
			email:              "test@example.com",
			password:           "password",
			createUserError:    dbstore.ErrUserExists,
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte("An account with this email already exists"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			userStore := &storemock.UserStoreMock{}

			userStore.On("CreateUser", tc.email, tc.password).Return(tc.createUserError)

			handler := NewPostRegisterHandler(PostRegisterHandlerParams{
				UserStore: userStore,
			})
			body := bytes.NewBufferString("email=" + tc.email + "&password=" + tc.password)
			req, _ := http.NewRequest("POST", "/", body)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			nonces := middleware.Nonces{
				Htmx:            "nonce-1234",
				ResponseTargets: "nonce-5678",
				Tw:              "nonce-9101",
				HtmxCSSHash:     "sha256-pgn1TCGZX6O77zDvy0oTODMOxemn0oj0LeCnQTRj7Kg=",
			}
			req = req.WithContext(context.WithValue(req.Context(), middleware.NonceKey, nonces))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(tc.expectedStatusCode, rr.Code, "handler returned wrong status code: got %v want %v", rr.Code, tc.expectedStatusCode)

			assert.True(bytes.Contains(rr.Body.Bytes(), tc.expectedBody), "handler returned unexpected body: got %v want %v", rr.Body.String(), tc.expectedBody)

			userStore.AssertExpectations(t)
		})
	}
}
