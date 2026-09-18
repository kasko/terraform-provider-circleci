package circleci

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.Handler) *ApiClient {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	origV2 := defaultV2BaseURL
	origWait := initialRetryWait
	u, _ := url.Parse(srv.URL + "/api/v2/")
	defaultV2BaseURL = u
	initialRetryWait = time.Millisecond
	t.Cleanup(func() {
		defaultV2BaseURL = origV2
		initialRetryWait = origWait
	})

	v1, _ := url.Parse(srv.URL + "/api/v1.1/")
	return &ApiClient{BaseURL: v1, Token: "secret"}
}

func TestGetProject_found(t *testing.T) {
	var gotPath, gotToken, gotQuery string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotToken = r.Header.Get("Circle-Token")
		w.Write([]byte(`{"id":"11111111-2222-4333-8444-555555555555","slug":"gh/kasko/zurich-svc","name":"zurich-svc","organization_id":"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"}`))
	}))

	p, err := c.GetProject("github", "kasko", "zurich-svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.VcsType != "github" || p.Username != "kasko" || p.Reponame != "zurich-svc" {
		t.Fatalf("unexpected project: %+v", p)
	}
	if p.ID != "11111111-2222-4333-8444-555555555555" || p.OrganizationID != "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee" || p.Slug != "gh/kasko/zurich-svc" {
		t.Fatalf("unexpected v2 attributes: %+v", p)
	}
	if gotPath != "/api/v2/project/gh/kasko/zurich-svc" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotToken != "secret" || gotQuery != "" {
		t.Fatalf("token must be sent as header only, got header=%q query=%q", gotToken, gotQuery)
	}
}

func TestGetProject_notFound(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Project not found"}`))
	}))

	_, err := c.GetProject("github", "kasko", "missing-svc")
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestGetProject_retriesOn429(t *testing.T) {
	var calls int32
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{}`))
	}))

	if _, err := c.GetProject("github", "kasko", "zurich-svc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestGetProject_otherErrorIsNotNotFound(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"Permission denied"}`))
	}))

	_, err := c.GetProject("github", "kasko", "zurich-svc")
	if err == nil || errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected a non-not-found error, got %v", err)
	}
}
