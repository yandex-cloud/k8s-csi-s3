# Helm chart for csi-s3

This chart adds S3 volume support to your cluster.

## Install chart

- Helm 2.x: `helm install [--set secret.accessKey=... --set secret.secretKey=... ...] --namespace kube-system --name csi-s3 .`
- Helm 3.x: `helm install [--set secret.accessKey=... --set secret.secretKey=... ...] --namespace kube-system csi-s3 .`

After installation succeeds, you can get a status of Chart: `helm status csi-s3`.

## Delete Chart

- Helm 2.x: `helm delete --purge csi-s3`
- Helm 3.x: `helm uninstall csi-s3 --namespace kube-system`

## Configuration

By default, this chart creates a secret and a storage class. You should at least set `secret.accessKey` and `secret.secretKey`
to your [Yandex Object Storage](https://cloud.yandex.com/en-ru/services/storage) keys for it to work.

The following table lists all configuration parameters and their default values.

| Parameter                           | Description                                                            | Default                                                |
| ----------------------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------ |
| `storageClass.create`               | Specifies whether the storage class should be created                  | true                                                   |
| `storageClass.name`                 | Storage class name                                                     | csi-s3                                                 |
| `storageClass.singleBucket`         | Use a single bucket for all dynamically provisioned persistent volumes |                                                        |
| `storageClass.mounter`              | Mounter to use. Either geesefs, s3fs or rclone. geesefs recommended    | geesefs                                                |
| `storageClass.mountOptions`         | GeeseFS mount options                                                  | `--memory-limit 1000 --dir-mode 0777 --file-mode 0666` |
| `storageClass.reclaimPolicy`        | Volume reclaim policy                                                  | Delete                                                 |
| `storageClass.annotations`          | Annotations for the storage class                                      |                                                        |
| `secret.create`                     | Specifies whether the secret should be created                         | true                                                   |
| `secret.name`                       | Name of the secret                                                     | csi-s3-secret                                          |
| `secret.accessKey`                  | S3 Access Key                                                          |                                                        |
| `secret.secretKey`                  | S3 Secret Key                                                          |                                                        |
| `secret.endpoint`                   | Endpoint                                                               | https://storage.yandexcloud.net                        |
| `secret.region`                     | Region                                                                 |                                                        |
| `tolerations.all`                   | Tolerate all taints by the CSI-S3 node driver (mounter)                | false                                                  |
| `tolerations.node`                  | Custom tolerations for the CSI-S3 node driver (mounter)                | []                                                     |
| `tolerations.controller`            | Custom tolerations for the CSI-S3 controller (provisioner)             | []                                                     |
| `initContainer.enabled`             | Enable init container for node daemon set                              | false                                                  |
| `initContainer.image`               | Init container image for node daemon set                               |                                                        |
| `initContainer.imagePullPolicy`     | Init container image pull policy for node daemon set                   | IfNotPresent                                           |
| `initContainer.command`             | Init container command for node daemon set                             | []                                                     |
| `initContainer.args`                | Init container args for node daemon set                                | []                                                     |
| `initContainer.volumeMounts`        | Init container volume mounts for node daemon set                       |                                                        |
| `initContainerProvisioner.enabled`  | Enable init container for provisioner stateful set                     | false                                                  |
| `initContainerProvisioner.image`    | Init container image for provisioner stateful set                      |                                                        |
| `initContainerProvisioner.imagePullPolicy` | Init container image pull policy for provisioner           | IfNotPresent                                           |
| `initContainerProvisioner.command`  | Init container command for provisioner stateful set                    | []                                                     |
| `initContainerProvisioner.args`     | Init container args for provisioner stateful set                       | []                                                     |
| `initContainerProvisioner.volumeMounts` | Init container volume mounts for provisioner                       |                                                        |
| `extraVolumes`                      | Extra volumes to mount on all containers                               | []                                                     |
| `extraVolumeMounts`                 | Extra volume mounts for node daemon set containers                     | []                                                     |
| `extraVolumeMountsProvisioner`      | Extra volume mounts for provisioner stateful set containers            | []                                                     |
