package cluster

import (
	"context"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Deletion struct {
	ProjectId string
	ServiceName string
}

func (d *Deletion) AuthenticateCluster() (kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()

	if err != nil {
		return kubernetes.Clientset{}, err
	}

	client,err := kubernetes.NewForConfig(config)

	if err != nil {
		return kubernetes.Clientset{}, err
	}

	return *client, nil
}

func (d *Deletion) DeleteService() error {
	ctx := context.Background()

	client, err := d.AuthenticateCluster()

	if err != nil {
		return err
	}

	serviceName := strings.ReplaceAll(d.ServiceName, "\"", "")

	serviceDeleteErr := client.CoreV1().Services(d.ProjectId).Delete(ctx, serviceName, v1.DeleteOptions{})

	if serviceDeleteErr != nil {
		return serviceDeleteErr
	}

	deploymentDeleteErr := client.AppsV1().Deployments(d.ProjectId).Delete(ctx, serviceName, v1.DeleteOptions{})

	if deploymentDeleteErr != nil {
		return deploymentDeleteErr
	}

	return nil
}

func (d *Deletion) DeleteEnvironment() error {
	ctx := context.Background()

	client, err := d.AuthenticateCluster()

	if err != nil {
		return err
	}

	deleteErr := client.CoreV1().Namespaces().Delete(ctx, d.ProjectId, v1.DeleteOptions{})

	if deleteErr != nil {
		return deleteErr
	}

	return nil
}
