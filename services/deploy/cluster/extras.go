package cluster

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Extras struct {
	ImageName string
}

func (e *Extras) Env(env map[string]string) []corev1.EnvVar {	
	// allocate len(env) + 1 to add the APP_NAME variable
	vars := make([]corev1.EnvVar, len(env) + 1) 

	vars[0] = corev1.EnvVar{
		Name: "APP_NAME",
		Value: e.ImageName,
	}

	for key, value := range env {
		vars = append(vars, corev1.EnvVar{
			Name: key,
			Value: value,
		})
	}

	return vars
}

type Ports struct {
	ContainerPort int `json:"container_port"`
	Name string `json:"name"`
}
func (e *Extras) Ports(ports []Ports) []corev1.ContainerPort {
	var containerPorts []corev1.ContainerPort

	containerPorts = append(containerPorts, corev1.ContainerPort{
		Name: "logger",
		Protocol: corev1.ProtocolTCP,
		ContainerPort: 5000,
	})

	for _, port := range ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name: port.Name,
			ContainerPort: int32(port.ContainerPort),
			Protocol: corev1.ProtocolTCP,
		})
	} 

	return containerPorts
}

func (e *Extras) ServicePorts(ports []Ports) []corev1.ServicePort {
	var servicePorts []corev1.ServicePort

	for _, port := range ports {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name: port.Name,
			Port: int32(port.ContainerPort),
			TargetPort: intstr.FromInt(port.ContainerPort),
			Protocol: corev1.ProtocolTCP,
		})
	}

	fmt.Println(servicePorts)

	return servicePorts
}
