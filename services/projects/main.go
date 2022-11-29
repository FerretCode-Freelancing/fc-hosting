package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

		data := dsnap.Data()
		stringified, err := json.Marshal(data)

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
		unescaped, err := url.QueryUnescape(name)

		if err != nil {
			http.Error(w, "There was an error fetching the name of your project.", http.StatusBadRequest)

			return
		}

		dsnap := client.Collection("users").Doc(fmt.Sprint(userId))

		doc, err := dsnap.Get(ctx)

		if err != nil {
			http.Error(w, "There was an error creating the project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		type Projects struct {
			Projects []map[string]interface{} `json:"projects"`
		}
		data := doc.Data()

		var projects Projects

		bytes, err := json.Marshal(data)

		if err != nil {
			http.Error(w, "There was an error creating the project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		if err := json.Unmarshal(bytes, &projects); err != nil {
			http.Error(w, "There was an error creating the project.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}
		
		dsnap.Update(ctx, []firestore.Update{
			{
				Path: "projects",
			},
		})
	})
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