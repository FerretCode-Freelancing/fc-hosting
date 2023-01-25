package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	events "github.com/ferretcode-freelancing/fc-bus"
	"github.com/ferretcode-freelancing/upload/projects"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/kubemq-io/kubemq-go"
)

type User struct {
	OwnerId   int    `json:"owner_id"`
	OwnerName string `json:"owner_name"`
}

type UploadRequest struct {
	RepoUrl   string            `json:"repo_url"`
	ProjectId string            `json:"project_id"`
	Ports     []Port            `json:"ports"`
	Env       map[string]string `json:"env"`
}

type DeleteRequest struct {
	ProjectId string `json:"project_id"`
}

type DeleteMessage struct {
	ServiceName string `json:"service_name"`
	ProjectId   string `json:"project_id"`
	Operation   string `json:"operation"`
}

type BuildMessage struct {
	RepoUrl   string            `json:"repo_url"`
	ProjectId string            `json:"project_id"`
	OwnerName string            `json:"owner_name"`
	Cookie    string            `json:"cookie"`
	Ports     []Port            `json:"ports"`
	Env       map[string]string `json:"env"`
	RamLimit  string            `json:"ram_limit"`
}

type Port struct {
	ContainerPort int    `json:"container_port"`
	Name          string `json:"name"`
}

func main() {
	ctx := context.Background()

	bus := events.Bus{
		Channel:       "build-pipeline",
		ClientId:      uuid.NewString(),
		Context:       ctx,
		TransportType: kubemq.TransportTypeGRPC,
	}

	client, err := bus.Connect()

	if err != nil {
		log.Fatalf("There was an error connecting to the message bus: %s\n", err)

		return
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Delete("/api/upload/delete", func(w http.ResponseWriter, r *http.Request) {
		_, _, ok := Authenticate(w, r)

		if !ok {
			return
		}

		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "There was an error deleting this resource!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		dr := DeleteRequest{}

		if err := json.Unmarshal(body, &dr); err != nil {
			http.Error(w, "There was an error deleting this resource!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		serviceName := r.URL.Query().Get("service_name")

		message := DeleteMessage{
			ProjectId: dr.ProjectId,
			Operation: "delete",
		}

		if len(serviceName) > 0 {
			message.ServiceName = serviceName
		}

		stringified, err := json.Marshal(message)

		if err != nil {
			http.Error(w, "There was an error deleting the selected resource!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		_, sendErr := client.Send(ctx, kubemq.NewQueueMessage().
			SetId(uuid.NewString()).
			SetChannel("deploy-pipeline").
			SetBody(stringified))

		if sendErr != nil {
			http.Error(w, "There was an error deleting the selected resource", http.StatusInternalServerError)

			return
		}

		w.WriteHeader(202)
		w.Write([]byte("Your repository was set to be deleted and will be deleted soon."))

	})

	r.Post("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		user, cookie, ok := Authenticate(w, r)

		if !ok {
			return
		}

		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "There was an error uploading your repository!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		ur := UploadRequest{}

		if err := json.Unmarshal(body, &ur); err != nil {
			http.Error(
				w,
				"Failed to get the information about the repository! You might need to add the repo_url field in your JSON body.",
				http.StatusInternalServerError,
			)

			fmt.Println(err)

			return
		}

		project, err := projects.GetProject(ur.ProjectId)

		if err != nil {
			http.Error(w, "Failed to validate your subscription!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		if project.SubscriptionId == "" {
			http.Error(w, "You do not have an active subscription!", http.StatusForbidden)

			return
		}

		message := BuildMessage{
			RepoUrl:   ur.RepoUrl,
			OwnerName: user.OwnerName,
			ProjectId: ur.ProjectId,
			Cookie:    cookie,
			Ports:     ur.Ports,
			Env:       ur.Env,
			RamLimit:  project.RamLimit,
		}

		fmt.Println(message)

		stringified, err := json.Marshal(message)

		if err != nil {
			http.Error(w, "There was an error uploading your repository.", http.StatusInternalServerError)

			return
		}

		_, sendErr := client.Send(ctx, kubemq.NewQueueMessage().
			SetId(uuid.NewString()).
			SetChannel(bus.Channel).
			SetBody(stringified))

		if sendErr != nil {
			http.Error(w, "There was an error uploading your repository", http.StatusInternalServerError)

			return
		}

		w.WriteHeader(202)
		w.Write([]byte("Your repository was uploaded successfully and is now building!"))
	})

	http.ListenAndServe(":3000", r)
}

func Authenticate(w http.ResponseWriter, r *http.Request) (User, string, bool) {
	client := &http.Client{}

	auth := fmt.Sprintf(
		"http://%s:%s",
		os.Getenv("FC_AUTH_SERVICE_HOST"),
		os.Getenv("FC_AUTH_SERVICE_PORT"),
	)

	if r.Body == nil {
		http.Error(w, "The body cannot be empty!", http.StatusBadRequest)

		return User{}, "", false
	}

	userReq, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/github/user", auth), nil)

	if err != nil {
		http.Error(w, "Failed to validate auth!", http.StatusInternalServerError)

		fmt.Println(err)

		return User{}, "", false
	}

	cookie, err := r.Cookie("fc-hosting")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		fmt.Println(err)

		return User{}, "", false
	}

	userReq.AddCookie(cookie)

	userReq.Close = true

	res, err := client.Do(userReq)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		fmt.Println(err)

		return User{}, "", false
	}

	if res.StatusCode != 200 {
		http.Error(w, "You are not authenticated!", http.StatusForbidden)

		fmt.Println(err)

		return User{}, "", false
	}

	defer res.Body.Close()

	authBody, err := io.ReadAll(res.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		fmt.Println(err)

		return User{}, "", false
	}

	var user User

	if err := json.Unmarshal(authBody, &user); err != nil {
		http.Error(w, "Failed to get the current user information!", http.StatusInternalServerError)

		fmt.Println(err)

		return User{}, "", false
	}

	return user, cookie.Value, true
}
