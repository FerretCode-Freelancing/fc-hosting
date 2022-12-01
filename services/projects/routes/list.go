package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ferretcode-freelancing/fc-hosting/projects/auth"
)

func List(w http.ResponseWriter, r *http.Request) {
	app, ctx, err := auth.InitializeFirebase()

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

	userId, err := auth.GetUserId(w, r)

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
		Projects        map[string]interface{} `json:"projects"`
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
}
