terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
      version = ">= 2.0.0"
    }
  }
}

provider "kubernetes" {
	config_path = "/root/.k3d/kubeconfig-fc-hosting.yaml"
}

resource "kubernetes_deployment" "default" {
	metadata {
		name = "registry"
		labels = {
      app = "registry"			
		}
	}

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "registry"
      }
    }

    template {
      metadata {
        labels = {
          app = "registry"
        }
      }

      spec {
        container {
          image = "registry:2"
          name = "registry"
          port {
            container_port = 5000
          }

          # env
          env {
            name = "REGISTRY_HTTP_TLS_CERTIFICATE"
            value = "/certs/tls.crt"
          }
          env {
            name = "REGISTRY_HTTP_TLS_KEY"
            value = "/certs/tls.key"
          }

          volume_mount {
            name = "registry-certs"
            mount_path = "/certs"
            read_only = true
          }

          volume_mount {
            name = "registry"
            mount_path = "/var/lib/registry"
            sub_path = "registry"
          }
        } 

        volume {
          name = "registry-certs"
          secret {
            secret_name = "registry-cert"
          }
        }

        volume {
          name = "registry"
          persistent_volume_claim {
            claim_name = "registry-pvc"
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "default" {
  metadata {
    name = "registry"
  }

  spec {
    selector = {
      app = "registry"
    }
    type = "LoadBalancer"
    port {
      port = 5000
      target_port = 5000
    }
    load_balancer_ip = "10.211.55.250"
  }
}

resource "kubernetes_persistent_volume_claim" "default" {
  metadata {
    name = "registry-pvc"
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    storage_class_name = "local-path"
    resources {
      requests = {
        storage = "5Gi"
      }
    }
  }
}
