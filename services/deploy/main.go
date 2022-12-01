package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ferretcode-freelancing/fc-hosting/services/deploy/cluster"
)

type DeployRequest struct {
	Ports []cluster.Ports `json:"ports"`
	Env map[string]string `json:"env"`
}
type Project struct {
	ImageName string `json:"imageName"`
}
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

		// fetch project to get imageName
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("%s/api/projects/get?id=%s", projects, projectId),
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

		// begin deploy process by getting extra env variables or ports
		var deployRequest DeployRequest

		processErr := ProcessBody(&deployRequest, r) 

		if processErr != nil {
			http.Error(w, "There was an error deploying the supplied repository!", http.StatusInternalServerError)	

			fmt.Println(err)

			return
		}

		deployment := cluster.Deployment{
			ImageName: project.ImageName,
			NamespaceName: project.ImageName,
			Extras: cluster.Extras{
				ImageName: project.ImageName,
			},
			Ports: deployRequest.Ports,
			Env: deployRequest.Env,
		}

		// apply resources to cluster
		deployErr := deployment.ApplyResources()

		if deployErr != nil {
			http.Error(w, "There was an error deploying the supplied repository.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}
	})

	http.ListenAndServe(":3000", r)
}

// s is the struct to unmarshal the body into
func ProcessBody(s interface{}, r *http.Request) error {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &r); err != nil {
		return err
	} 

	return nil
}
