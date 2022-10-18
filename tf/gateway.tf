resource "kubernetes_ingress_v1" "default" {
  metadata {
    name = "gateway"
  }

  spec {
    default_backend {
      service {
        name = "fc-gateway"
        port {
          number = 3001
        }
      }
    }
  }
}

