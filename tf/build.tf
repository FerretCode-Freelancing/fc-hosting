resource "kubernetes_deployment" "fc-build" {
  metadata {
    name = "fc-build"
    labels = {
      app = "fc-build"
    } 
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "fc-build"
      }
    }

    template {
      metadata {
        labels = {
          app = "fc-build"
        }
      }

      spec {
        container {
          image = "sthanguy/fc-build"
          name = "fc-build"
        }
      }
    }
  }
}
