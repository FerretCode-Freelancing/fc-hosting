package cluster

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Deletion struct {
	ProjectId string
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
