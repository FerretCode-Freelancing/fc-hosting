package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type User struct {
	OwnerId int `json:"owner_id"`
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
		client := &http.Client{}

		auth := fmt.Sprintf(
			"http://%s:%s",
			os.Getenv("FC_AUTH_SERVICE_HOST"),
			os.Getenv("FC_AUTH_SERVICE_PORT"),
		)

		authenticated := CheckSession(w, r)

		if authenticated != true {
			http.Error(w, "You are not authenticated!", http.StatusForbidden)

			return
		}

		if r.Body == nil {
			http.Error(w, "The body cannot be empty!", http.StatusBadRequest)

			return
		}

		userReq, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/github/user", auth), nil)

		if err != nil {
			http.Error(w, "Failed to validate auth!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		sid, err := r.Cookie("fc-hosting")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		cookie := &http.Cookie{
			Name: "fc-hosting",
			Value: sid.Value,
		}

		userReq.AddCookie(cookie)

		userReq.Close = true

		res, err := client.Do(userReq)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		if res.StatusCode != 200 {
			http.Error(w, "You are not authenticated!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		defer res.Body.Close()

		authBody, err := ioutil.ReadAll(res.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		var user User

		if err := json.Unmarshal(authBody, &user); err != nil {
			http.Error(w, "Failed to get the current user information!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		fmt.Println(user)

		var ur UploadRequest

		if err := json.Unmarshal(body, &ur); err != nil {
			http.Error(
				w,
				"Failed to get the information about the repository! You might need to add the repo_url field in your JSON body.", 
				http.StatusInternalServerError,
			)

			fmt.Println(err)

			return
		} 

		fmt.Println(ur)

		builder := fmt.Sprintf(
			"http://%s:%s@%s:%s",
			strings.Trim(os.Getenv("FC_BUILDER_USERNAME"), "\n"),
			strings.Trim(os.Getenv("FC_BUILDER_PASSWORD"), "\n"),
			os.Getenv("FC_PROVISION_SERVICE_HOST"),
			os.Getenv("FC_PROVISION_SERVICE_PORT"),
		)
		
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

			fmt.Println(err)

			return
		}

		resp, err := client.Do(req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		defer res.Body.Close()

		if resp.StatusCode == 200 {
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
		fmt.Sprintf("%s?sid=%s", cache, cookie.Value[4:36]), // cut string to 32 chars
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

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return true
	}

	if res.Body != nil {
		return true
	}

	return false
}
