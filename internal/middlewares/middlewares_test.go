package middlewares

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type writer struct {
	statusCode int
	sb         strings.Builder
}

func (w *writer) Header() http.Header {
	return nil
}

func (w *writer) Write(bytes []byte) (int, error) {
	return w.sb.Write(bytes)
}

func (w *writer) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func TestRecoverer(t *testing.T) {
	hPanic := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("try to recover me")
	})

	hNormal := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	wr := new(writer)

	assert.NotPanics(t, func() {
		Recoverer()(hPanic).ServeHTTP(wr, nil)
	}, "must not panic")

	Recoverer()(hNormal).ServeHTTP(wr, nil)
}
