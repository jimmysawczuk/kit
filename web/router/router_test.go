package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jimmysawczuk/kit/web/router"
)

func TestRouteParamsWithMiddleware(t *testing.T) {
	r := router.New()

	// Middleware that does nothing
	nopMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	var capturedID string

	// Register route with middleware - this triggers the bug
	r.Get("/users/{userID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should extract "123" but returns "" due to bug
		capturedID = chi.URLParam(r, "userID")
		w.WriteHeader(http.StatusOK)
	}), nopMiddleware)

	req := httptest.NewRequest("GET", "/users/123", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if capturedID != "123" {
		t.Errorf("expected userID to be '123', got '%s'", capturedID)
	}
}

func TestRouteParamsWithoutMiddleware(t *testing.T) {
	r := router.New()

	var capturedID string

	// Register route WITHOUT middleware - this works fine
	r.Get("/users/{userID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = chi.URLParam(r, "userID")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/users/456", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if capturedID != "456" {
		t.Errorf("expected userID to be '456', got '%s'", capturedID)
	}
}

type authModule struct{}

func (m authModule) Route(r router.Router) {
	r.Getf("/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Getf("/logout", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

func TestBind(t *testing.T) {
	r := router.New()
	r.Bind("/v1/auth", authModule{})

	for _, tt := range []struct {
		path string
		want int
	}{
		{"/v1/auth/login", http.StatusOK},
		{"/v1/auth/logout", http.StatusNoContent},
		{"/v1/auth/missing", http.StatusNotFound},
	} {
		req := httptest.NewRequest("GET", tt.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tt.want {
			t.Errorf("GET %s: expected %d, got %d", tt.path, tt.want, rec.Code)
		}
	}
}

func TestBindWithMiddleware(t *testing.T) {
	r := router.New()

	var middlewareCalled bool
	trackMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	}

	r.Bind("/v1/auth", authModule{}, trackMiddleware)

	req := httptest.NewRequest("GET", "/v1/auth/login", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !middlewareCalled {
		t.Error("expected middleware to be called")
	}
}

func TestNestedRouteParamsWithMiddleware(t *testing.T) {
	r := router.New()

	nopMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	var capturedSiteID, capturedRecipeID string

	// Test nested routes like /v1/recipes/{siteID}/recipes/{recipeID}
	r.Route("/v1/recipes", func(r router.Router) {
		r.Get("/{siteID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedSiteID = chi.URLParam(r, "siteID")
			w.WriteHeader(http.StatusOK)
		}), nopMiddleware)

		r.Get("/{siteID}/{recipeID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedSiteID = chi.URLParam(r, "siteID")
			capturedRecipeID = chi.URLParam(r, "recipeID")
			w.WriteHeader(http.StatusOK)
		}), nopMiddleware)
	})

	// Test single param
	req1 := httptest.NewRequest("GET", "/v1/recipes/site-abc", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec1.Code)
	}
	if capturedSiteID != "site-abc" {
		t.Errorf("expected siteID to be 'site-abc', got '%s'", capturedSiteID)
	}

	// Test multiple params
	capturedSiteID, capturedRecipeID = "", ""
	req2 := httptest.NewRequest("GET", "/v1/recipes/site-xyz/recipe-789", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec2.Code)
	}
	if capturedSiteID != "site-xyz" {
		t.Errorf("expected siteID to be 'site-xyz', got '%s'", capturedSiteID)
	}
	if capturedRecipeID != "recipe-789" {
		t.Errorf("expected recipeID to be 'recipe-789', got '%s'", capturedRecipeID)
	}
}
