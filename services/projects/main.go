package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	events "github.com/ferretcode-freelancing/fc-bus"
	"github.com/ferretcode-freelancing/fc-hosting/projects/cache"
	"github.com/ferretcode-freelancing/fc-hosting/projects/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/kubemq-io/kubemq-go"
)

type StatusUpdate struct {
	ProjectId string `json:"project_id"`
	ServiceName string `json:"service_name"`
	NewStatus string `json:"new_status"`
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	cache := cache.NewCache()

	r.Get("/api/projects/list", func(w http.ResponseWriter, r *http.Request) {
		routes.List(w, r)
	})

	r.Get("/api/projects/status", func(w http.ResponseWriter, r *http.Request) {
		routes.Status(w, r, cache)
	})

	r.Post("/api/projects/new", func(w http.ResponseWriter, r *http.Request) {
		routes.New(w, r)
	})

	r.Get("/api/projects/get", func(w http.ResponseWriter, r *http.Request) {
		routes.Get(w, r)
	})

	r.Patch("/api/projects/update", func(w http.ResponseWriter, r *http.Request) {
		routes.Update(w, r)
	})

	go http.ListenAndServe(":3000", r)

	ctx := context.Background()

	bus := events.Bus{
		Channel: "status-updates",
		ClientId: uuid.NewString(),
		Context: ctx,
		TransportType: kubemq.TransportTypeGRPC,
	}

	client, err := bus.Connect()

	if err != nil {
		log.Fatal("There was an error connecting to the message bus.")
	}

	done, err := bus.Subscribe(client, func(msgs *kubemq.ReceiveQueueMessagesResponse, subscribeError error) {
		if subscribeError != nil {
			log.Printf("There was an error processing a message: %s\n", subscribeError.Error())

			message := msgs.Messages[0]

			fmt.Printf("Message recieved: %s\n", string(message.Body))

			update := StatusUpdate{}

			if err := json.Unmarshal(message.Body, &update); err != nil {
				log.Printf("There was an error processing a status update: %s\n", subscribeError.Error())
			}

			cache.AddStatus(update.ProjectId, update.ServiceName, update.NewStatus)
		}
	})

	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGTERM)
	signal.Notify(shutdown, syscall.SIGINT)
	signal.Notify(shutdown, syscall.SIGQUIT)

	select {
	case <-shutdown:
		done <- struct{}{}
	}

}
