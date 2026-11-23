package handlers

import (
	"bytes"
	"goth/internal/store"
	storemock "goth/internal/store/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUsersPage(t *testing.T) {
	admin := &store.User{ID: 1, Email: "boss@example.com"}

	t.Run("admin sees users with actions", func(t *testing.T) {
		assert := assert.New(t)
		userStore := &storemock.UserStoreMock{}
		userStore.On("ListUsers").Return([]store.User{
			{ID: 1, Email: "boss@example.com", IsAdmin: true},
			{ID: 2, Email: "user@example.com", IsAdmin: false},
		}, nil)

		handler := NewUsersHandler(UsersHandlerParams{
			UserStore: userStore,
			LogStore:  &storemock.LogStoreMock{},
		})

		req, _ := http.NewRequest("GET", "/admin/users", nil)
		req = req.WithContext(adminTestContext(admin, true))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		for _, expected := range []string{"Users", "boss@example.com", "user@example.com", "Admin", "Make admin"} {
			assert.True(bytes.Contains(rr.Body.Bytes(), []byte(expected)), "body missing %q", expected)
		}
		assert.False(bytes.Contains(rr.Body.Bytes(), []byte("/admin/users/1/revoke")), "own row must not offer revoke")
		userStore.AssertExpectations(t)
	})

	t.Run("non-admin redirects home", func(t *testing.T) {
		assert := assert.New(t)
		handler := NewUsersHandler(UsersHandlerParams{})

		req, _ := http.NewRequest("GET", "/admin/users", nil)
		req = req.WithContext(adminTestContext(&store.User{ID: 2, Email: "user@example.com"}, false))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusFound, rr.Code)
		assert.Equal("/", rr.Header().Get("Location"))
	})

	t.Run("anonymous redirects to login", func(t *testing.T) {
		assert := assert.New(t)
		handler := NewUsersHandler(UsersHandlerParams{})

		req, _ := http.NewRequest("GET", "/admin/users", nil)
		req = req.WithContext(adminTestContext(nil, false))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusFound, rr.Code)
		assert.Equal("/login", rr.Header().Get("Location"))
	})
}

func TestUsersGrantRevoke(t *testing.T) {
	admin := &store.User{ID: 1, Email: "boss@example.com"}

	newDeleteRequest := func(target string, user *store.User, isAdmin bool) *http.Request {
		req, _ := http.NewRequest("POST", target, nil)
		rctx := routeCtx(map[string]string{"id": "2"})
		return req.WithContext(withRouteCtx(adminTestContext(user, isAdmin), rctx))
	}

	t.Run("grant redirects to users", func(t *testing.T) {
		assert := assert.New(t)
		userStore := &storemock.UserStoreMock{}
		logStore := &storemock.LogStoreMock{}
		userStore.On("SetAdmin", uint(2), true).Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		handler := NewUsersHandler(UsersHandlerParams{UserStore: userStore, LogStore: logStore})

		rr := httptest.NewRecorder()
		handler.GrantAdmin(rr, newDeleteRequest("/admin/users/2/grant", admin, true))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Equal("/admin/users", rr.Header().Get("Location"))
		userStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("revoke redirects to users", func(t *testing.T) {
		assert := assert.New(t)
		userStore := &storemock.UserStoreMock{}
		logStore := &storemock.LogStoreMock{}
		userStore.On("SetAdmin", uint(2), false).Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		handler := NewUsersHandler(UsersHandlerParams{UserStore: userStore, LogStore: logStore})

		rr := httptest.NewRecorder()
		handler.RevokeAdmin(rr, newDeleteRequest("/admin/users/2/revoke", admin, true))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Equal("/admin/users", rr.Header().Get("Location"))
		userStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("non-admin cannot grant", func(t *testing.T) {
		assert := assert.New(t)
		userStore := &storemock.UserStoreMock{}

		handler := NewUsersHandler(UsersHandlerParams{UserStore: userStore, LogStore: &storemock.LogStoreMock{}})

		rr := httptest.NewRecorder()
		handler.GrantAdmin(rr, newDeleteRequest("/admin/users/2/grant", &store.User{ID: 2, Email: "user@example.com"}, false))

		assert.Equal(http.StatusFound, rr.Code)
		userStore.AssertExpectations(t)
	})
}
