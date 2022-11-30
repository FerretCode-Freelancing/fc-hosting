package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)

	r.Post("/deploy", func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{}

		projects := fmt.Sprintf(
			"http://%s:%s",
			os.Getenv("FC_PROJECTS_SERVICE_HOST"),
			os.Getenv("FC_PROJECTS_SERVICE_PORT"),
		)

		projectId := r.URL.Query().Get("id")

		if projectId == "" {
			http.Error(w, "You need to supply a project ID!", http.StatusBadRequest)

			return
		}

		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("%s/api/projects/get", projects),
			nil,
		)

		if err != nil {
			http.Error(w, "There was an error fetching the supplied project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		res, err := client.Do(req)

		if err != nil {
			http.Error(w, "There was an error fetching the supplied project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		type Project struct {
			ImageName string `json:"imageName"`
		}

		body, err := io.ReadAll(res.Body)

		if err != nil {
			http.Error(w, "There was an error deploying the supplied project.", http.StatusInternalServerError)

			fmt.Println(err)

			return	
		}

		var project Project

		if err := json.Unmarshal(body, &project); err != nil {
			http.Error(w, "There was an error deploying the supplied project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}
	})

	http.ListenAndServe(":3000", r)
}
