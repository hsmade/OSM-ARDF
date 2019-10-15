package web

import (
	"github.com/matryer/is"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRootRedirect(t *testing.T) {
	is := is.New(t)
	srv := NewServer()
	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)
	is.Equal(w.Result().StatusCode, http.StatusMovedPermanently)
	is.Equal(w.Header().Get("location"), "/app/")
}
