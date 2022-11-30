package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	r.Get("/api/projects/list", func(w http.ResponseWriter, r *http.Request) {
		app, ctx, err := InitializeFirebase()

		if err != nil {
			fmt.Println(err)

			http.Error(w, "There was an issue fetching all projects!", http.StatusInternalServerError)

			return
		}

		client, err := app.Firestore(ctx)

		if err != nil {
			fmt.Println(err)

			http.Error(w, "There was an issue fetching all projects!", http.StatusInternalServerError)

			return
		}

		userId, err := GetUserId(w, r)

		if err != nil {
			if err.Error() == "you are not authenticated" {
				http.Error(w, "You are not authenticated!", http.StatusForbidden)

				return
			}

			fmt.Println(err)

			http.Error(w, "There was an error authenticating you!", http.StatusInternalServerError)

			return
		}

		dsnap, err := client.Collection("users").Doc(fmt.Sprint(userId)).Get(ctx)

		if err != nil {
			http.Error(w, "There was an issue fetching your projects!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		type Projects struct {
			Projects map[string]interface{} `json:"projects"`
			RunningProjects map[string]interface{} `json:"runningProjects"`
		}
		data := dsnap.Data()
		stringifiedRaw, err := json.Marshal(data)

		if err != nil {
			http.Error(w, "There was an issue fetching your projects!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		var projects Projects
		if err := json.Unmarshal(stringifiedRaw, &projects); err != nil {
			http.Error(w, "There was an issue fetching your projects!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		stringified, err := json.Marshal(projects)

		if err != nil {
			http.Error(w, "There was an issue fetching your projects!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		w.WriteHeader(200)
		w.Write(stringified)
	})
	
	r.Post("/api/projects/new", func(w http.ResponseWriter, r *http.Request) {
		app, ctx, err := InitializeFirebase()
		
		if err != nil {
			http.Error(w, "There was an issue creating your project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		client, err := app.Firestore(ctx)
		 
		if err != nil {
			http.Error(w, "There was an issue creating your project.", http.StatusInternalServerError)

			fmt.Println(err)

			return 
		}

		userId, err := GetUserId(w, r)

		if err != nil {
			http.Error(w, "There was an issue fetching the current user.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		name := r.URL.Query().Get("name")

		if err != nil {
			http.Error(w, "There was an error fetching the name of your project.", http.StatusBadRequest)

			return
		}

		dsnap := client.Collection("users").Doc(fmt.Sprint(userId))

		res, err := dsnap.Set(ctx, map[string]interface{}{
			"projects": map[string]interface{}{
				uuid.New().String(): map[string]interface{}{
					"name": name,
				},
			},
		}, firestore.MergeAll)

		if err != nil {
			http.Error(w, "There was an error creating the project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		fmt.Println(res)

		w.WriteHeader(200)
		w.Write([]byte("The project was created successfully."))
	})

	r.Get("/api/projects/get", func(w http.ResponseWriter, r *http.Request) {
		app, ctx, err := InitializeFirebase()

		if err != nil {
			http.Error(w, "There was an error fetching the project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		client, err := app.Firestore(ctx)

		if err != nil {
			http.Error(w, "There was an error fetching the project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		userId, err := GetUserId(w, r)

		if err != nil {
			http.Error(w, "There was an error authenticating you.", http.StatusForbidden)

			fmt.Println(err)

			return
		}

		dsnap, err := client.Collection("users").Doc(fmt.Sprint(userId)).Get(ctx)

		if err != nil {
			http.Error(w, "There was an error fetching the project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		if !dsnap.Exists() {
			http.Error(w, "The provided project does not exist.", http.StatusNotFound)

			return
		}

		data := dsnap.Data()

		projectId := r.URL.Query().Get("id")
		stringified, err := json.Marshal(
			data["projects"].(map[string]interface{})[projectId],
		)

		if err != nil {
			http.Error(w, "The project was found but there was an error processing it.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		w.WriteHeader(200)
		w.Write(stringified)
	})

	r.Patch("/api/projects/update", func(w http.ResponseWriter, r *http.Request) {
		app, ctx, err := InitializeFirebase()

		if err != nil {
			http.Error(w, "There was an error updating the project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		client, err := app.Firestore(ctx)

		if err != nil {
			http.Error(w, "There was an error parsing the request body.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "There was an error parsing the request body.", http.StatusBadRequest)

			fmt.Println(err)

			return
		}

		updates := make(map[string]interface{})
		projectId := r.URL.Query().Get("id")

		if err := json.Unmarshal(body, &updates); err != nil {
			http.Error(w, "There was an error updating the project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		userId, err := GetUserId(w, r)

		if err != nil {
			http.Error(w, "There was an error authenticating you!", http.StatusForbidden)

			fmt.Println(err)

			return
		}

		dsnap := client.Collection("users").Doc(fmt.Sprint(userId))

		if _, err := dsnap.Set(ctx, map[string]interface{}{
			"projects": map[string]interface{}{
				projectId: updates,
			},
		}, firestore.MergeAll); err != nil {
			http.Error(w, "There was an issue updating the project!", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		w.WriteHeader(200)
		w.Write([]byte("The doc was successfully updated."))
	})

	http.ListenAndServe(":3000", r)
}

func InitializeFirebase() (firebase.App, context.Context, error) {
	ctx := context.Background()

	app, err := firebase.NewApp(ctx, nil)

	if err != nil {
		return firebase.App{}, ctx, err
	}
	
	return *app, ctx, nil
}

type User struct {
	OwnerId int `json:"owner_id"`
}
func GetUserId(w http.ResponseWriter, r *http.Request) (int, error) {
	client := &http.Client{}

	auth := fmt.Sprintf(
		"http://%s:%s",
		os.Getenv("FC_AUTH_SERVICE_HOST"),
		os.Getenv("FC_AUTH_SERVICE_PORT"),
	)

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/auth/github/user", auth),
		nil,
	)
	
	if err != nil {
		return 0, err
	}

	cookie, err := r.Cookie("fc-hosting")

	if err != nil {
		return 0, err
	}

	req.AddCookie(cookie)

	req.Close = true

	res, err := client.Do(req)

	if err != nil {
		return 0, err
	}

	if res.StatusCode != 200 {
		return 0, errors.New("you are not authenticated")
	}

	userBody, err := io.ReadAll(res.Body)

	if err != nil {
		return 0, err
	}

	var user User 

	if err := json.Unmarshal(userBody, &user); err != nil {
		return 0, err
	}

	return user.OwnerId, nil
}
