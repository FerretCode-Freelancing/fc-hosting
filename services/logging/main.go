package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Request struct {
	Webhook bool   `json:"webhook"`
	Message string `json:"message"`
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	file, err := os.Open("./config/custom/secret")

	if err != nil {
		fmt.Println("Couldn't start logger")
		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	scanner.Scan()

	webhook := scanner.Text()

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		request := Request{}

		err := json.NewDecoder(r.Body).Decode(&request)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		if request.Webhook {
			request.Message = fmt.Sprintf(
				"%s - %s",
				time.Now().Local().Format("RFC822"),
				request.Message,
			)

			body := []byte(fmt.Sprintf(`{"content": "%s"}`, request.Message))

			req, err := http.NewRequest(
				"POST",
				string(webhook),
				bytes.NewBuffer(body),
			)
			sendError(err, w)

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			res, err := client.Do(req)
			sendError(err, w)

			defer res.Body.Close()

			w.Write([]byte(`Webhook sent`))

			return
		}

		fmt.Printf(
			"%s - %s\n",
			time.Now().UTC(),
			request.Message,
		)

		w.Write([]byte(`OK`))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(""))
	})

	http.ListenAndServe(":5000", r)
}

func sendError(e error, w http.ResponseWriter) {
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}
