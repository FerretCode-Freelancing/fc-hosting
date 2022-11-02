resource "kubernetes_deployment" "fc-provision" {
  metadata {
    name = "fc-provision"
    labels = {
      app = "fc-provision"
    } 
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "fc-provision"
      }
    }

    template {
      metadata {
        labels = {
          app = "fc-provision"
        }
      }

      spec {
        container {
          image = "sthanguy/fc-provision"
          name = "fc-provision"
          port {
            container_port = 3000
          }

          env {
            name = "FC_BUILDER_USERNAME"
            value_from {
              secret_key_ref {
                name = "builder-secret"
                key = "username"
                optional = false
              }
            }
          }

          env {
            name = "FC_BUILDER_PASSWORD"
            value_from {
              secret_key_ref {
                name = "builder-secret"
                key = "password"
                optional = false
              }
            }
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

resource "kubernetes_service" "fc-provision" {
  metadata {
    name = "fc-provision"
  }

  spec {
    selector = {
      app = "fc-provision"
    }

    type = "ClusterIP"

    port {
      port = 3000
    }
  }
}
