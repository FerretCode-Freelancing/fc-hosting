package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ferretcode-freelancing/fc-hosting/projects/auth"
)

func Get(w http.ResponseWriter, r *http.Request) {
	app, ctx, err := auth.InitializeFirebase()

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

	userId, err := auth.GetUserId(w, r)

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

	if projectId == "" {
		http.Error(w, "The project ID needs to be set in the URL parameters.", http.StatusBadRequest)

		return
	}

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
}
