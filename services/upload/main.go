package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type User struct {
	OwnerId string `json:"owner_id"`
	OwnerName string `json:"owner_name"`
}

type UploadRequest struct {
	RepoUrl string `json:"repo_url"`
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

		if err := json.Unmarshal(authBody, &user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		var ur UploadRequest

		if err := json.Unmarshal(body, &ur); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} 

		builder := fmt.Sprintf(
			"http://%s:%s@%s:%s",
			strings.Trim(os.Getenv("FC_BUILDER_CACHE_USERNAME"), "\n"),
			strings.Trim(os.Getenv("FC_BUILDER_CACHE_PASSWORD"), "\n"),
			os.Getenv("FC_BUILDER_SERVICE_HOST"),
			os.Getenv("FC_BUILDER_SERVICE_PORT"),
		)

		client := &http.Client{}

		sid, err := r.Cookie("fc-hosting")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		req, err := http.NewRequest(
			"POST",
			builder,
			bytes.NewReader([]byte(
				fmt.Sprintf(
					`{ "repo_name": %s, "owner_name": %s, "cookie": %s }`,
					ur.RepoUrl,
					user.OwnerName,
					sid.Value,	
				),
			)),
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		res, err := client.Do(req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if res.StatusCode == 200 {
			w.WriteHeader(200)
			w.Write([]byte("Your repository was deployed successfully!"))
		}
	})

	http.ListenAndServe(":3000", r)
}

func CheckSession(w http.ResponseWriter, r *http.Request) bool {
	cache := fmt.Sprintf(
		"http://%s:%s@%s:%s", 
		strings.Trim(os.Getenv("FC_SESSION_CACHE_USERNAME"), "\n"),
		strings.Trim(os.Getenv("FC_SESSION_CACHE_PASSWORD"), "\n"),
		os.Getenv("FC_SESSION_CACHE_SERVICE_HOST"), 
		os.Getenv("FC_SESSION_CACHE_SERVICE_PORT"),
	)

	cookie, err := r.Cookie("fc-hosting")

	if err != nil {
		fmt.Println(err)

		return false
	}

	client := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s?sid=%s", cache, url.QueryEscape(cookie.Value)),
		nil,
	)

	if err != nil {
		fmt.Println(err)

		return false
	}

	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err)

		return false
	}

	if res.StatusCode == 200 {
		fmt.Println(err)

		return true
	}

	return false
}
