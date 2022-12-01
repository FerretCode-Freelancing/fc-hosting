package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/ferretcode-freelancing/fc-hosting/projects/auth"
)

func Update(w http.ResponseWriter, r *http.Request) {
	app, ctx, err := auth.InitializeFirebase()

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

	userId, err := auth.GetUserId(w, r)

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
}
