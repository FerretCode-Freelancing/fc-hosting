package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type User struct {
	UUID string `json:"id"`
}

type Env struct {
	Env string `json:"env"`
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		authenticated := CheckSession(w, r)

		if authenticated != true {
			http.Redirect(w, r, "http://localhost:3001/auth/github", http.StatusFound)

			return
		}

		if r.Body == nil {
			http.Error(w, "The body cannot be empty!", http.StatusBadRequest)

			return
		}
	})

	r.Post("/api/upload/env", func(w http.ResponseWriter, r *http.Request) {
		authenticated := CheckSession(w, r)

		if authenticated != true {
			http.Error(w, errors.New("You are not authenticated!").Error(), http.StatusForbidden)

			return
		}

		if r.Body == nil {
			http.Error(w, "The body cannot be empty!", http.StatusBadRequest)

			return
		}

		resp, err := http.Get("http://localhost:3001/auth/github/user")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		authBody, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		var user User

		if err = json.Unmarshal(authBody, &user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		var env Env

		if err = json.Unmarshal(body, &env); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		//TODO: write internal service for creating deployments
	})

	http.ListenAndServe(":3000", r)
}

func CheckSession(w http.ResponseWriter, r *http.Request) bool {
	cache := fmt.Sprintf("%s:%s", os.Getenv("FC_SESSION_SESSION_CACHE_HOST"), os.Getenv("FC_SESSION_CACHE_SERVICE_PORT"))

	cookie, err := r.Cookie("fc-hosting")

	if err != nil {
		return false
	}

	client := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s?sid=%s", cache, cookie.Value),
		nil,
	)

	if err != nil {
		return false
	}

	res, err := client.Do(req)

	if err != nil {
		return false
	}

	if res.StatusCode == 200 {
		return true
	}

	return false
}
