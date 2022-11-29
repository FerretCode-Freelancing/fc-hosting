package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

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

		fmt.Println(data)
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