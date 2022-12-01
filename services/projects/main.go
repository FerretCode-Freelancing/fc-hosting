package main

import (
	"net/http"

	"github.com/ferretcode-freelancing/fc-hosting/projects/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	r.Get("/api/projects/list", func(w http.ResponseWriter, r *http.Request) {
		routes.List(w, r)
	})

	r.Post("/api/projects/new", func(w http.ResponseWriter, r *http.Request) {
		routes.New(w, r)
	})

	r.Get("/api/projects/get", func(w http.ResponseWriter, r *http.Request) {
		routes.Get(w, r)
	})

	r.Patch("/api/projects/update", func(w http.ResponseWriter, r *http.Request) {
		routes.Update(w, r)
	})

	http.ListenAndServe(":3000", r)
}
