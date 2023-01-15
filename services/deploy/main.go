package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ferretcode-freelancing/fc-hosting/services/deploy/cluster"
	"github.com/ferretcode-freelancing/fc-bus"
	"github.com/google/uuid"
	"github.com/kubemq-io/kubemq-go"
)

type DeployRequest struct {
	Ports []cluster.Ports `json:"ports"`
	Env map[string]string `json:"env"`
	ImageName string `json:"image_name"`
	ProjectId string `json:"project_id"`
	Operation string `json:"operation"`
}

func main() {
	ctx := context.Background()

	bus := events.Bus{
		Channel: "deploy-pipeline",	
		ClientId: uuid.NewString(),
		Context: ctx,
		TransportType: kubemq.TransportTypeGRPC,
	}

	fmt.Println(bus.ClientId)

	client, err := bus.Connect()

	if err != nil {
		log.Fatal("There was an error connecting to the message bus.")
	}

	done, err := bus.Subscribe(client, func(msgs *kubemq.ReceiveQueueMessagesResponse, subscribeError error) {
		if subscribeError != nil {
			log.Printf("There was an error processing a message: %s\n", subscribeError.Error())

			return
		}

		message := msgs.Messages[0]
		request := &DeployRequest{}

		_, err := resolveService("fc-deploy.default.svc.cluster.local")

		if err != nil {
			log.Printf("There was an error processing a message: %s\n", subscribeError.Error())

			return
		}

		if err := json.Unmarshal(message.Body, &request); err != nil {
			log.Printf("There was an error deploying a project: %s", err)

			return
		} 

		if request.Operation == "create" {
			deployment := cluster.Deployment{
				ImageName: request.ImageName,
				NamespaceName: request.ProjectId,
				Extras: cluster.Extras{
					ImageName: request.ImageName,
				},
				Ports: request.Ports,
				Env: request.Env,
			}

			deployErr := deployment.ApplyResources()

			if deployErr != nil {
				fmt.Printf("There was an error applying resources: %s\n", deployErr)
			}

			return
		}

		deletion := cluster.Deletion{
			ProjectId: request.ProjectId,
		}

		deleteErr := deletion.DeleteEnvironment()

		if deleteErr != nil {
			log.Printf("There was an error deleting a project: %s", deleteErr)

			return
		}
})

	if err != nil {
		log.Println("There was an error subscribing to the queue.")
	}

	var shutdown = make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGTERM)
	signal.Notify(shutdown, syscall.SIGINT)
	signal.Notify(shutdown, syscall.SIGQUIT)

	select {
	case <-shutdown:
		done <- struct{}{}
	}
}

func resolveService(host string) (string, error) {
	ip, err := net.LookupIP(host)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", ip[0].String(), "3000"), nil
}

// s is the struct to unmarshal the body into
func ProcessBody(s interface{}, r *http.Request) error {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &r); err != nil {
		return err
	} 

	return nil
}
