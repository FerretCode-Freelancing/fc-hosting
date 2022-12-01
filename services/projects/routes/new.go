package routes

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/ferretcode-freelancing/fc-hosting/projects/auth"
	"github.com/google/uuid"
)

func New(w http.ResponseWriter, r *http.Request) {
	app, ctx, err := auth.InitializeFirebase()

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

	userId, err := auth.GetUserId(w, r)

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
}
