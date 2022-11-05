terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
      version = ">= 2.0.0"
    }
  }
}

provider "kubernetes" {
	config_path = "~/.kube/config"
}

resource "kubernetes_deployment" "fc-registry" {
	metadata {
		name = "fc-registry"
		labels = {
      app = "fc-registry"			
		}
	}

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "fc-registry"
      }
    }

    template {
      metadata {
        labels = {
          app = "fc-registry"
        }
      }

      spec {
        container {
          image = "registry:2"
          name = "fc-registry"
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
            name = "fc-registry"
            mount_path = "/var/lib/fc-registry"
            sub_path = "fc-registry"
          }
        } 

        volume {
          name = "registry-certs"
          secret {
            secret_name = "registry-cert"
          }
        }

        volume {
          name = "fc-registry"
          persistent_volume_claim {
            claim_name = "registry-pvc"
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "fc-registry" {
  metadata {
    name = "fc-registry"
  }

  spec {
    selector = {
      app = "fc-registry"
    }
    type = "ClusterIP"
    port {
      port = 5000
    }
  }
}

resource "kubernetes_persistent_volume_claim" "fc-registry" {
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
