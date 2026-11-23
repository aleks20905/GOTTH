package handlers

import (
	"errors"
	hashmock "goth/internal/hash/mock"
	"goth/internal/store"
	storemock "goth/internal/store/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostAccountEmail(t *testing.T) {
	user := &store.User{ID: 1, Email: "old@example.com", Password: "hash"}

	post := func(form url.Values, user *store.User) *httptest.ResponseRecorder {
		userStore := &storemock.UserStoreMock{}
		logStore := &storemock.LogStoreMock{}
		if form.Get("email") == "new@example.com" {
			userStore.On("UpdateEmail", uint(1), "new@example.com").Return(nil)
			logStore.On("AddEntry", mock.Anything).Return(nil)
		} else if form.Get("email") == "taken@example.com" {
			userStore.On("UpdateEmail", uint(1), "taken@example.com").Return(errors.New("taken"))
		}

		handler := NewPostAccountHandler(PostAccountHandlerParams{
			UserStore:    userStore,
			PasswordHash: &hashmock.PasswordHashMock{},
			LogStore:     logStore,
		})

		req, _ := http.NewRequest("POST", "/account/email", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		if user != nil {
			req = req.WithContext(adminTestContext(user, false))
		}

		rr := httptest.NewRecorder()
		handler.ChangeEmail(rr, req)

		userStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
		return rr
	}

	t.Run("valid email redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		rr := post(url.Values{"email": {"new@example.com"}}, user)
		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "msg=email-updated")
	})

	t.Run("taken email redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		rr := post(url.Values{"email": {"taken@example.com"}}, user)
		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "msg=email-taken")
	})

	t.Run("invalid email redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		rr := post(url.Values{"email": {"not-an-email"}}, user)
		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "msg=email-invalid")
	})

	t.Run("anonymous user redirects to login", func(t *testing.T) {
		assert := assert.New(t)
		rr := post(url.Values{"email": {"anon@example.com"}}, nil)
		assert.Equal(http.StatusFound, rr.Code)
		assert.Equal("/login", rr.Header().Get("Location"))
	})
}

func TestPostAccountPassword(t *testing.T) {
	user := &store.User{ID: 1, Email: "old@example.com", Password: "hash"}

	post := func(current string, next string, ctxUser *store.User, stub func(userStore *storemock.UserStoreMock, passwordHash *hashmock.PasswordHashMock, logStore *storemock.LogStoreMock)) *httptest.ResponseRecorder {
		userStore := &storemock.UserStoreMock{}
		passwordHash := &hashmock.PasswordHashMock{}
		logStore := &storemock.LogStoreMock{}
		if stub != nil {
			stub(userStore, passwordHash, logStore)
		}

		handler := NewPostAccountHandler(PostAccountHandlerParams{
			UserStore:    userStore,
			PasswordHash: passwordHash,
			LogStore:     logStore,
		})

		form := url.Values{"current_password": {current}, "new_password": {next}}
		req, _ := http.NewRequest("POST", "/account/password", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		if ctxUser != nil {
			req = req.WithContext(adminTestContext(ctxUser, false))
		}

		rr := httptest.NewRecorder()
		handler.ChangePassword(rr, req)

		userStore.AssertExpectations(t)
		passwordHash.AssertExpectations(t)
		logStore.AssertExpectations(t)
		return rr
	}

	t.Run("valid change redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		rr := post("oldpass12", "newpass34", user, func(userStore *storemock.UserStoreMock, passwordHash *hashmock.PasswordHashMock, logStore *storemock.LogStoreMock) {
			passwordHash.On("ComparePasswordAndHash", "oldpass12", "hash").Return(true, nil)
			passwordHash.On("GenerateFromPassword", "newpass34").Return("newhash", nil)
			userStore.On("UpdatePassword", uint(1), "newhash").Return(nil)
			logStore.On("AddEntry", mock.Anything).Return(nil)
		})
		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "msg=password-updated")
	})

	t.Run("wrong current password redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		rr := post("wrongpass", "newpass34", user, func(userStore *storemock.UserStoreMock, passwordHash *hashmock.PasswordHashMock, logStore *storemock.LogStoreMock) {
			passwordHash.On("ComparePasswordAndHash", "wrongpass", "hash").Return(false, nil)
		})
		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "msg=password-wrong")
	})

	t.Run("short password redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		rr := post("oldpass12", "short", user, nil)
		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "msg=password-short")
	})

	t.Run("anonymous user redirects to login", func(t *testing.T) {
		assert := assert.New(t)
		rr := post("oldpass12", "newpass34", nil, nil)
		assert.Equal(http.StatusFound, rr.Code)
		assert.Equal("/login", rr.Header().Get("Location"))
	})
}

func TestGetAccountPage(t *testing.T) {
	t.Run("page shows email", func(t *testing.T) {
		assert := assert.New(t)
		handler := NewAccountHandler()

		req, _ := http.NewRequest("GET", "/account", nil)
		req = req.WithContext(adminTestContext(&store.User{ID: 1, Email: "me@example.com"}, false))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		assert.Contains(rr.Body.String(), "me@example.com")
		assert.Contains(rr.Body.String(), "Change email")
		assert.Contains(rr.Body.String(), "Change password")
	})

	t.Run("anonymous user redirects to login", func(t *testing.T) {
		assert := assert.New(t)
		handler := NewAccountHandler()

		req, _ := http.NewRequest("GET", "/account", nil)
		req = req.WithContext(adminTestContext(nil, false))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusFound, rr.Code)
		assert.Equal("/login", rr.Header().Get("Location"))
	})
}
