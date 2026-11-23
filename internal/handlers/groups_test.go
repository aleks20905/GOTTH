package handlers

import (
	"bytes"
	"errors"
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

func TestGroupsPage(t *testing.T) {
	admin := &store.User{ID: 1, Email: "boss@example.com"}

	t.Run("page renders groups and devices", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		groupStore := &storemock.GroupStoreMock{}
		groupStore.On("ListGroups").Return([]store.DeviceGroup{
			{ID: 1, Name: "Finance", Description: "Accounting PCs"},
		}, nil)
		deviceStore.On("ListDevices").Return([]store.Device{
			{ID: 1, Name: "finance-pc"},
		}, nil)

		handler := NewGroupsHandler(GroupsHandlerParams{
			DeviceStore: deviceStore,
			GroupStore:  groupStore,
			LogStore:    &storemock.LogStoreMock{},
		})

		req, _ := http.NewRequest("GET", "/admin/groups", nil)
		req = req.WithContext(adminTestContext(admin, true))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		for _, expected := range []string{"Groups", "Finance", "finance-pc"} {
			assert.True(bytes.Contains(rr.Body.Bytes(), []byte(expected)), "body missing %q", expected)
		}
		groupStore.AssertExpectations(t)
		deviceStore.AssertExpectations(t)
	})

	t.Run("non-admin redirects home", func(t *testing.T) {
		assert := assert.New(t)
		handler := NewGroupsHandler(GroupsHandlerParams{})

		req, _ := http.NewRequest("GET", "/admin/groups", nil)
		req = req.WithContext(adminTestContext(&store.User{ID: 2, Email: "user@example.com"}, false))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusFound, rr.Code)
		assert.Equal("/", rr.Header().Get("Location"))
	})
}

func TestGroupsMutations(t *testing.T) {
	admin := &store.User{ID: 1, Email: "boss@example.com"}

	newHandler := func() (*GroupsHandler, *storemock.DeviceStoreMock, *storemock.GroupStoreMock, *storemock.LogStoreMock) {
		deviceStore := &storemock.DeviceStoreMock{}
		groupStore := &storemock.GroupStoreMock{}
		logStore := &storemock.LogStoreMock{}
		handler := NewGroupsHandler(GroupsHandlerParams{
			DeviceStore: deviceStore,
			GroupStore:  groupStore,
			LogStore:    logStore,
		})
		return handler, deviceStore, groupStore, logStore
	}

	post := func(target string, form url.Values) *http.Request {
		req, _ := http.NewRequest("POST", target, strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		return req.WithContext(adminTestContext(admin, true))
	}

	t.Run("add group redirects", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, groupStore, logStore := newHandler()
		groupStore.On("AddGroup", mock.MatchedBy(func(g *store.DeviceGroup) bool {
			return g.Name == "Finance"
		})).Return(&store.DeviceGroup{ID: 1}, nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		rr := httptest.NewRecorder()
		handler.AddGroup(rr, post("/admin/groups", url.Values{"name": {"Finance"}}))

		assert.Equal(http.StatusSeeOther, rr.Code)
		groupStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("empty name redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, _, _ := newHandler()

		rr := httptest.NewRecorder()
		handler.AddGroup(rr, post("/admin/groups", url.Values{"name": {""}}))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "error=empty")
	})

	t.Run("duplicate name redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, groupStore, _ := newHandler()
		groupStore.On("AddGroup", mock.Anything).Return(nil, errors.New("UNIQUE constraint failed"))

		rr := httptest.NewRecorder()
		handler.AddGroup(rr, post("/admin/groups", url.Values{"name": {"Finance"}}))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "error=exists")
		groupStore.AssertExpectations(t)
	})

	t.Run("delete group redirects", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, groupStore, logStore := newHandler()
		groupStore.On("DeleteGroup", uint(1)).Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		req, _ := http.NewRequest("POST", "/admin/groups/1/delete", nil)
		rctx := routeCtx(map[string]string{"id": "1"})
		req = req.WithContext(withRouteCtx(adminTestContext(admin, true), rctx))

		rr := httptest.NewRecorder()
		handler.DeleteGroup(rr, req)

		assert.Equal(http.StatusSeeOther, rr.Code)
		groupStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("assign and unassign redirect", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, _, logStore := newHandler()
		gid := uint(1)
		deviceStore.On("AssignGroup", uint(2), &gid).Return(nil)
		deviceStore.On("AssignGroup", uint(2), (*uint)(nil)).Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		req, _ := http.NewRequest("POST", "/admin/groups/1/assign", strings.NewReader(url.Values{"device_id": {"2"}}.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(withRouteCtx(adminTestContext(admin, true), routeCtx(map[string]string{"id": "1"})))

		rr := httptest.NewRecorder()
		handler.AssignDevice(rr, req)
		assert.Equal(http.StatusSeeOther, rr.Code)

		req2, _ := http.NewRequest("POST", "/admin/groups/1/unassign/2", nil)
		req2 = req2.WithContext(withRouteCtx(adminTestContext(admin, true), routeCtx(map[string]string{"id": "1", "deviceID": "2"})))

		rr = httptest.NewRecorder()
		handler.UnassignDevice(rr, req2)
		assert.Equal(http.StatusSeeOther, rr.Code)

		deviceStore.AssertExpectations(t)
	})
}
