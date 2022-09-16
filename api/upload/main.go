package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
)

// This struct provides information on how to cut session ID strings
// to check with Redis
// As of now the defaults are
/*
	Start: 4
	End: 36
	Prefix: sess:
*/
type SessionString struct {
	Start  int
	End    int
	Prefix string
}

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
	session := SessionString{
		Start:  4,
		End:    36,
		Prefix: "sess:",
	}

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("FC_SESSION_STORAGE_SERVICE_HOST"), os.Getenv("FC_SESSION_STORAGE_SERVICE_PORT")),
	})

	cookie, err := r.Cookie("fc-hosting")
	sendError(w, err)

	_, redisNotFound := rdb.Get(
		ctx,
		fmt.Sprintf("%s%s", session.Prefix, cookie.Value[session.Start:session.End]),
	).Result()

	if redisNotFound != nil {
		http.Redirect(
			w,
			r,
			"http://localhost:3001/auth/github",
			http.StatusForbidden,
		)

		return false
	}

	return true
}

func sendError(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	return
}
