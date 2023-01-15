package cluster

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	fmt.Println(d.ImageName)

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
							Image: fmt.Sprintf(
								"%s:%s/%s", 
								strings.Trim(os.Getenv("FC_REGISTRY_SERVICE_HOST"), "\n"),
								strings.Trim(os.Getenv("FC_REGISTRY_SERVICE_PORT"), "\n"),
								d.ImageName,
							),
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
			Type: corev1.ServiceTypeLoadBalancer,
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

	fmt.Println("Authenticated")

	// create resource definitions
	namespace := d.CreateNamespace()
	deployment := d.CreateDeployment()
	service := d.CreateService()

	fmt.Println("Schemas created")

	namespaceExists, err := d.CheckNamespace(namespace.Name, ctx, client)

	if err != nil {
		return err
	}

	serviceExists, err := d.CheckDeploymentAndService(deployment.Name, namespace.Name, ctx, client)

	if err != nil {
		return err
	}

	// apply resources
	if !namespaceExists {
		_, namespaceCreateErr := client.CoreV1().Namespaces().Create(ctx, &namespace, v1.CreateOptions{})
		if namespaceCreateErr != nil {
			return namespaceCreateErr
		}
	}
	
	if !serviceExists {
		_, deploymentCreateErr := client.AppsV1().Deployments(d.NamespaceName).Create(ctx, &deployment, v1.CreateOptions{})
		if deploymentCreateErr != nil {
			return deploymentCreateErr 
		}

		_, serviceCreateErr := client.CoreV1().Services(d.NamespaceName).Create(ctx, &service, v1.CreateOptions{})
		if serviceCreateErr != nil {
			return serviceCreateErr
		}

		fmt.Println("Definitions created")

		return nil
	}

	_, deploymentUpdateErr := client.AppsV1().Deployments(d.NamespaceName).Update(ctx, &deployment, v1.UpdateOptions{})
	if deploymentUpdateErr != nil {
		return deploymentUpdateErr
	}

	_, serviceUpdateErr := client.CoreV1().Services(d.NamespaceName).Update(ctx, &service, v1.UpdateOptions{})
	if serviceUpdateErr != nil {
		return serviceUpdateErr
	}

	return nil
}

func (d *Deployment) CheckDeploymentAndService(
	name string,
	namespace string,
	ctx context.Context,
	client kubernetes.Clientset,
) (bool, error) {
	deploymentList, err := client.AppsV1().Deployments(namespace).List(ctx, v1.ListOptions{})
	serviceList, err := client.CoreV1().Services(namespace).List(ctx, v1.ListOptions{})

	if err != nil {
		return true, err
	}

	for _, deployment := range deploymentList.Items {
		if deployment.Name == name {
			return true, nil
		}
	}

	for _, service := range serviceList.Items {
		if service.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (d *Deployment) CheckNamespace(
	name string, 
	ctx context.Context, 
	client kubernetes.Clientset,
) (bool, error) {
	list, err := client.CoreV1().Namespaces().List(ctx, v1.ListOptions{})	

	if err != nil {
		return true, err
	}

	for _, namespace := range list.Items {
		fmt.Println(namespace.Name)
		fmt.Println(name)

		if namespace.Name == name {
			return true, nil
		} 
	}

	return false, nil
}
