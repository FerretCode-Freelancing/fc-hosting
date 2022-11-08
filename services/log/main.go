package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)

	r.Get("/api/log/{deployment}", func(w http.ResponseWriter, r *http.Request) {
		deployment := chi.URLParam(r, "deployment")
	})
}
