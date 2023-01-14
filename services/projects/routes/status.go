package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ferretcode-freelancing/fc-hosting/projects/cache"
)

func Status(w http.ResponseWriter, r *http.Request, cache cache.Cache) {
	projectId := r.URL.Query().Get("id")

	if projectId == "" {
		http.Error(w, "The project ID needs to be set in the URL parameters.", http.StatusBadRequest)

		return
	}

	statuses, exists := cache.Get(projectId)

	if !exists {
		http.Error(w, "There is no known status information for this project yet.", http.StatusNotFound)

		return
	}

	status := r.URL.Query().Get("service")

	if status == "" {
		stringified, err := json.Marshal(
			statuses[status],
		)

		if err != nil {
			http.Error(w, "There was an error processing the status.", http.StatusInternalServerError)

			fmt.Println(err)

			return
		}

		w.WriteHeader(200)
		w.Write(stringified)
	}

	stringified, err := json.Marshal(
		statuses,
	)

	if err != nil {
		http.Error(w, "There was an error processing the statuses.", http.StatusInternalServerError)

		fmt.Println(err)

		return
	}

	w.WriteHeader(200)
	w.Write(stringified)
}
