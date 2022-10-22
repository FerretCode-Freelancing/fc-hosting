resource "kubernetes_deployment" "fc-session-cache" {
  metadata {
    name = "fc-session-cache"
    labels = {
      app = "fc-session-cache"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "fc-session-cache"
      }
    }

    template {
      metadata {
        labels = {
          app = "fc-session-cache"
        } 
      }

      spec {
        container {
          image = "sthanguy/fc-session-cache"
          name = "fc-session-cache"
          port {
            container_port = 3005
          }

          env {
            name = "FC_SESSION_CACHE_USERNAME"
            value_from {
              secret_key_ref {
                name = "session-cache-secret"
                key = "username"
                optional = false
              }
            }
          }

          env {
            name = "FC_SESSION_CACHE_PASSWORD"
            value_from {
              secret_key_ref {
                name = "session-cache-secret"
                key = "password"
                optional = false
              }
            }
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "fc-session-cache" {
  metadata {
    name = "fc-session-cache"
  }

  spec {
    selector = {
      app = "fc-session-cache"
    }
    
    type = "ClusterIP"

    port {
      port = 3000
    }
  }
}
