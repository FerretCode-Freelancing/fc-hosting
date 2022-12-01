package cluster

import (
	corev1 "k8s.io/api/core/v1"
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
	ContainerPort int `json:"containerPort"`
	Name string `json:"name"`
}
func (e *Extras) Ports(ports []Ports) []corev1.ContainerPort {
	// allocate len(ports) + 1 so we can add a port for the logger
	containerPorts := make([]corev1.ContainerPort, len(ports) + 1)	

	containerPorts[0] = corev1.ContainerPort{
		Name: "logger",
		Protocol: corev1.ProtocolTCP,
		ContainerPort: 5000,
	}

	for _, port := range ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name: port.Name,
			ContainerPort: int32(port.ContainerPort),
			Protocol: corev1.ProtocolTCP,
		})
	} 

	return containerPorts
}
