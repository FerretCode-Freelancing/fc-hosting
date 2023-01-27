package cluster

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
)

type Deployment struct {
	ImageName     string
	NamespaceName string
	Extras        Extras
	Ports         []Ports
	Env           map[string]string
	RamLimit      string
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

func (d *Deployment) CreateResourceQuota() (corev1.ResourceQuota, error) {
	cpuLimit, err := resource.ParseQuantity("1")

	if err != nil {
		return corev1.ResourceQuota{}, err
	}

	ramLimit, err := resource.ParseQuantity(d.RamLimit)

	if err != nil {
		return corev1.ResourceQuota{}, err
	}

	resourceQuota := corev1.ResourceQuota{
		ObjectMeta: v1.ObjectMeta{
			Name:      d.NamespaceName,
			Namespace: d.NamespaceName,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceLimitsCPU:    cpuLimit,
				corev1.ResourceLimitsMemory: ramLimit,
			},
		},
	}

	return resourceQuota, nil
}

func (d *Deployment) DeployStatus() appsv1.Deployment {
	deployment := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "fc-status",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "fc-status",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": "fc-status",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "fc-status",
							Image: "sthanguy/fc-status",
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment
}

func (d *Deployment) CreateDeployment() appsv1.Deployment {
	fmt.Println(d.Ports)

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
							Env:   d.Extras.Env(d.Env),
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
			Type:  corev1.ServiceTypeNodePort,
			Ports: d.Extras.ServicePorts(d.Ports),
			Selector: map[string]string{
				"app": d.ImageName,
			},
		},
	}

	return service
}

func (d *Deployment) CreateIngress(port int32) networking.Ingress {
	pathType := networking.PathType("Prefix")

	paths := networking.HTTPIngressPath{
		Path:     fmt.Sprintf("/%s/%s/", d.NamespaceName, d.ImageName),
		PathType: &pathType,
		Backend: networking.IngressBackend{
			Service: &networking.IngressServiceBackend{
				Name: d.ImageName,
				Port: networking.ServiceBackendPort{
					Number: port,
				},
			},
		},
	}

	ruleValue := networking.IngressRuleValue{
		HTTP: &networking.HTTPIngressRuleValue{
			Paths: []networking.HTTPIngressPath{paths},
		},
	}

	rules := networking.IngressRule{
		IngressRuleValue: ruleValue,
	}

	ingress := networking.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name: d.ImageName,
			Labels: map[string]string{
				"run": d.ImageName,
			},
			Annotations: map[string]string{
				"ingress.kubernetes.io/ssl-redirect": "false",
			},
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{rules},
		},
	}

	return ingress
}

func (d *Deployment) CreateRole() (rbac.Role, rbac.RoleBinding) {
	rule := rbac.PolicyRule{
		APIGroups: []string{""},
		Resources: []string{"pods"},
		Verbs:     []string{"list"},
	}

	role := rbac.Role{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-status", d.NamespaceName),
			Namespace: d.NamespaceName,
		},
		Rules: []rbac.PolicyRule{rule},
	}

	subject := rbac.Subject{
		Kind:      "ServiceAccount",
		Name:      "fc-status",
		Namespace: d.NamespaceName,
	}

	roleBinding := rbac.RoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-status", d.NamespaceName),
			Namespace: d.NamespaceName,
		},
		Subjects: []rbac.Subject{subject},
		RoleRef: rbac.RoleRef{
			Kind:     "Role",
			Name:     fmt.Sprintf("%s-status", d.NamespaceName),
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	return role, roleBinding
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

	namespaceExists, err := d.CheckNamespace(namespace.Name, ctx, client)

	if err != nil {
		return err
	}

	serviceExists, err := d.CheckDeploymentAndService(deployment.Name, namespace.Name, ctx, client)

	if err != nil {
		return err
	}

	// this is despicable

	// apply resources
	if !namespaceExists {
		err := d.CreateProjectResources(client, ctx)

		if err != nil {
			return err
		}
	}

	if !serviceExists {
		err := d.CreateProjectResources(client, ctx)

		if err != nil {
			return err
		}

		return nil
	}

	updateErr := d.UpdateProjectService(client, ctx)

	if updateErr != nil {
		return err
	}

	return nil
}

func (d *Deployment) UpdateProjectService(client kubernetes.Clientset, ctx context.Context) error {
	deployment := d.CreateDeployment()
	service := d.CreateService()
	ingress := d.CreateIngress(int32(d.Ports[0].ContainerPort))

	_, deploymentUpdateErr := client.AppsV1().Deployments(d.NamespaceName).Update(ctx, &deployment, v1.UpdateOptions{})
	if deploymentUpdateErr != nil {
		return deploymentUpdateErr
	}

	_, serviceUpdateErr := client.CoreV1().Services(d.NamespaceName).Update(ctx, &service, v1.UpdateOptions{})
	if serviceUpdateErr != nil {
		return serviceUpdateErr
	}

	_, ingressUpdateErr := client.NetworkingV1().Ingresses(d.NamespaceName).Update(ctx, &ingress, v1.UpdateOptions{})
	if ingressUpdateErr != nil {
		return ingressUpdateErr
	}

	patch := fmt.Sprintf(`{"spec": {
		"template": {"metadata": {
			"annotations": {
				"kubectl.kubernetes.io/restartedAt": "%s"
			}
		}}
	}}`, time.Now().Format(time.RFC3339))

	client.AppsV1().Deployments(d.NamespaceName).Patch(ctx, deployment.Name, types.StrategicMergePatchType, []byte(patch), v1.PatchOptions{})

	return nil

}

func (d *Deployment) CreateProjectService(client kubernetes.Clientset, ctx context.Context) error {
	deployment := d.CreateDeployment()
	service := d.CreateService()
	ingress := d.CreateIngress(int32(d.Ports[0].ContainerPort))

	_, deploymentCreateErr := client.AppsV1().Deployments(d.NamespaceName).Create(ctx, &deployment, v1.CreateOptions{})
	if deploymentCreateErr != nil {
		return deploymentCreateErr
	}

	_, serviceCreateErr := client.CoreV1().Services(d.NamespaceName).Create(ctx, &service, v1.CreateOptions{})
	if serviceCreateErr != nil {
		return serviceCreateErr
	}

	_, ingressCreateErr := client.NetworkingV1().Ingresses(d.NamespaceName).Create(ctx, &ingress, v1.CreateOptions{})
	if ingressCreateErr != nil {
		return ingressCreateErr
	}

	return nil
}

func (d *Deployment) CreateProjectResources(client kubernetes.Clientset, ctx context.Context) error {
	namespace := d.CreateNamespace()
	role, roleBinding := d.CreateRole()
	statusDeployment := d.DeployStatus()
	resourceQuota, err := d.CreateResourceQuota()

	if err != nil {
		return err
	}

	_, namespaceCreateErr := client.CoreV1().Namespaces().Create(ctx, &namespace, v1.CreateOptions{})
	if namespaceCreateErr != nil {
		return namespaceCreateErr
	}

	_, resourceQuotaCreateErr := client.CoreV1().ResourceQuotas(d.NamespaceName).Create(ctx, &resourceQuota, v1.CreateOptions{})
	if resourceQuotaCreateErr != nil {
		return resourceQuotaCreateErr
	}

	_, roleCreateErr := client.RbacV1().Roles(d.NamespaceName).Create(ctx, &role, v1.CreateOptions{})
	if roleCreateErr != nil {
		return roleCreateErr
	}

	_, roleBindingCreateErr := client.RbacV1().RoleBindings(d.NamespaceName).Create(ctx, &roleBinding, v1.CreateOptions{})
	if roleBindingCreateErr != nil {
		return roleBindingCreateErr
	}

	_, statusCreateErr := client.AppsV1().Deployments(d.NamespaceName).Create(ctx, &statusDeployment, v1.CreateOptions{})
	if statusCreateErr != nil {
		return statusCreateErr
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
		if namespace.Name == name {
			return true, nil
		}
	}

	return false, nil
}
