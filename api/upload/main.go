package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	file, err := os.Open("./config/session/secret")

	if err != nil {
		fmt.Println("Couldn't start upload service")
		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	scanner.Scan()

	// secret := scanner.Text()

	r.Get("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			fmt.Println("error fetching cookie")

			return
		}

		w.Write([]byte("hi"))
	})

	http.ListenAndServe(":3000", r)
}
