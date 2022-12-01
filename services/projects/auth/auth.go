package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
)

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
