job "plugin-s3-controller" {
  datacenters = ["dc1"]
  type        = "service"

  constraint {
    attribute = "${meta.type}"
    operator  = "set_contains_any"
    value     = "runner,sandbox,database"
  }

  group "controller" {
    count = 1

    task "plugin" {
      driver = "docker"

      config {
        image = "cr.yandex/crp9ftr22d26age3hulg/yandex-cloud/csi-s3/csi-s3-driver:0.43.7"

        args = [
          "--endpoint=unix:///csi/csi.sock",
          "--nodeid=controller-${node.unique.name}",
          "--logtostderr",
          "--v=5",
        ]
      }

      csi_plugin {
        id        = "s3"
        type      = "controller"
        mount_dir = "/csi"
      }

      resources {
        cpu    = 300
        memory = 301
      }
    }
  }
}
