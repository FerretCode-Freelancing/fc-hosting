package cluster

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
)

type Deployment struct {
	ImageName string
	NamespaceName string
	Extras Extras
	Ports []Ports
	Env map[string]string
}

func (d *Deployment) AuthenticateCluster() (kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()

	if err != nil {
		return kubernetes.Clientset{}, err
	}

	client, err := kubernetes.NewForConfig(config)

	if err != nil {
		return kubernetes.Clientset{}, err
	}

	return *client, nil
}

func (d *Deployment) CreateNamespace() corev1.Namespace {
	namespace := corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: d.NamespaceName,
		},
	}

	return namespace
}

func (d *Deployment) CreateDeployment() appsv1.Deployment {
	deployment := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: d.ImageName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": d.ImageName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": d.ImageName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: d.ImageName,
							Image: d.ImageName,
							Ports: d.Extras.Ports(d.Ports),
							Env: d.Extras.Env(d.Env),
						},
					},
				},
			},
		},
	}	

	return deployment
} 

func (d *Deployment) CreateService() corev1.Service {
	service := corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: d.ImageName,
			Labels: map[string]string{
				"run": d.ImageName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 3000,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": d.ImageName,
			},
		},
	}

	return service
}

func (d *Deployment) ApplyResources() error {
	ctx := context.Background()

	client, err := d.AuthenticateCluster()

	if err != nil {
		return err
	}

	// create resource definitions
	namespace := d.CreateNamespace()
	deployment := d.CreateDeployment()
	service := d.CreateService()

	// apply resources
	_, namespaceCreateErr := client.CoreV1().Namespaces().Create(ctx, &namespace, v1.CreateOptions{})
	if namespaceCreateErr != nil {
		return err
	}

	_, deploymentCreateErr := client.AppsV1().Deployments(d.NamespaceName).Create(ctx, &deployment, v1.CreateOptions{})
	if deploymentCreateErr != nil {
		return err
	}

	_, serviceCreateErr := client.CoreV1().Services(d.NamespaceName).Create(ctx, &service, v1.CreateOptions{})
	if serviceCreateErr != nil {
		return err
	}

	return nil
}
