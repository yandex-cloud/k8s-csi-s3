type      = "csi"
id        = "s3-demo"
name      = "s3-demo"
plugin_id = "s3"

capability {
  # This driver supports multi-node multi-writer access.
  access_mode     = "multi-node-multi-writer"
  attachment_mode = "file-system"
}

capacity_min = "1GiB"
capacity_max = "10GiB"

parameters {
  mounter = "geesefs"
  # Add --no-systemd if host systemd integration is not available.
  options = "--memory-limit 1000 --dir-mode 0777 --file-mode 0666 --no-systemd"
  # bucket = "your-existing-bucket" # Optional. If omitted, bucket name is generated per volume.
}

secrets {
  accessKeyID     = "REPLACE_ME"
  secretAccessKey = "REPLACE_ME"
  endpoint        = "https://storage.yandexcloud.net"
  # region        = "eu-central-1" # Optional for non-AWS backends.
  # insecure      = "true"         # Optional for self-signed TLS endpoints.
}
