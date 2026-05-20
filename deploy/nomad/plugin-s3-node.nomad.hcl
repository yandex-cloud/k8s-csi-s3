job "plugin-s3-node" {
  datacenters = ["dc1"]
  type        = "system"
  node_pool   = "all"

  constraint {
    attribute = "${meta.type}"
    operator  = "set_contains_any"
    value     = "runner,sandbox,database"
  }

  group "nodes" {
    task "plugin" {
      driver = "docker"

      config {
        image = "cr.yandex/crp9ftr22d26age3hulg/yandex-cloud/csi-s3/csi-s3-driver:0.43.7"

        args = [
          "--endpoint=unix:///csi/csi.sock",
          "--nodeid=${node.unique.name}",
          "--logtostderr",
          "--v=5",
        ]

        # Node plugins must be privileged for bidirectional mount propagation.
        privileged = true
      }

      csi_plugin {
        id        = "s3"
        type      = "node"
        mount_dir = "/csi"
      }

      resources {
        cpu    = 300
        memory = 301
      }
    }
  }
}
