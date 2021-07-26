# CSI for S3

This is a Container Storage Interface ([CSI](https://github.com/container-storage-interface/spec/blob/master/spec.md)) for S3 (or S3 compatible) storage. This can dynamically allocate buckets and mount them via a fuse mount into any container.

## Status

This is still very experimental and should not be used in any production environment. Unexpected data loss could occur depending on what mounter and S3 storage backend is being used.

## Kubernetes installation

### Requirements

* Kubernetes 1.13+ (CSI v1.0.0 compatibility)
* Kubernetes has to allow privileged containers
* Docker daemon must allow shared mounts (systemd flag `MountFlags=shared`)

### 1. Create a secret with your S3 credentials

```yaml
apiVersion: v1
kind: Secret
metadata:
  namespace: kube-system
  name: csi-s3-secret
  # Namespace depends on the configuration in the storageclass.yaml
  namespace: kube-system
stringData:
  accessKeyID: <YOUR_ACCESS_KEY_ID>
  secretAccessKey: <YOUR_SECRET_ACCES_KEY>
  # For AWS set it to "https://s3.<region>.amazonaws.com"
  endpoint: <S3_ENDPOINT_URL>
  # If not on S3, set it to ""
  region: <S3_REGION>
```

The region can be empty if you are using some other S3 compatible storage.

### 2. Deploy the driver

```bash
cd deploy/kubernetes
kubectl create -f provisioner.yaml
kubectl create -f attacher.yaml
kubectl create -f csi-s3.yaml
```

### 3. Create the storage class

```bash
kubectl create -f examples/storageclass.yaml
```

### 4. Test the S3 driver

1. Create a pvc using the new storage class:

    ```bash
    kubectl create -f examples/pvc.yaml
    ```

1. Check if the PVC has been bound:

    ```bash
    $ kubectl get pvc csi-s3-pvc
    NAME         STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
    csi-s3-pvc   Bound     pvc-c5d4634f-8507-11e8-9f33-0e243832354b   5Gi        RWO            csi-s3         9s
    ```

1. Create a test pod which mounts your volume:

    ```bash
    kubectl create -f examples/pod.yaml
    ```

    If the pod can start, everything should be working.

1. Test the mount

    ```bash
    $ kubectl exec -ti csi-s3-test-nginx bash
    $ mount | grep fuse
    s3fs on /var/lib/www/html type fuse.s3fs (rw,nosuid,nodev,relatime,user_id=0,group_id=0,allow_other)
    $ touch /var/lib/www/html/hello_world
    ```

If something does not work as expected, check the troubleshooting section below.

## Additional configuration

### Bucket

By default, csi-s3 will create a new bucket per volume. The bucket name will match that of the volume ID. If you want your volumes to live in a precreated bucket, you can simply specify the bucket in the storage class parameters:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: csi-s3-existing-bucket
provisioner: ch.ctrox.csi.s3-driver
parameters:
  mounter: rclone
  bucket: some-existing-bucket-name
```

If the bucket is specified, it will still be created if it does not exist on the backend. Every volume will get its own prefix within the bucket which matches the volume ID. When deleting a volume, also just the prefix will be deleted.

### Mounter

As S3 is not a real file system there are some limitations to consider here. Depending on what mounter you are using, you will have different levels of POSIX compability. Also depending on what S3 storage backend you are using there are not always [consistency guarantees](https://github.com/gaul/are-we-consistent-yet#observed-consistency).

The driver can be configured to use one of these mounters to mount buckets:

* [geesefs](https://github.com/yandex-cloud/geesefs) (recommended and default)
* [s3fs](https://github.com/s3fs-fuse/s3fs-fuse)
* [rclone](https://rclone.org/commands/rclone_mount)

The mounter can be set as a parameter in the storage class. You can also create multiple storage classes for each mounter if you like.

Characteristics of different mounters (for more detailed information consult their own documentation):

#### GeeseFS

* Almost full POSIX compatibility
* Good performance for both small and big files
* Files can be viewed normally with any S3 client

#### s3fs

* Almost full POSIX compatibility
* Good performance for big files, poor performance for small files
* Files can be viewed normally with any S3 client

#### rclone

* Less POSIX compatible than s3fs
* Bad performance for big files, okayish performance for small files
* Files can be viewed normally with any S3 client
* Doesn't create directory objects like s3fs or GeeseFS

## Troubleshooting

### Issues while creating PVC

Check the logs of the provisioner:

```bash
kubectl logs -l app=csi-provisioner-s3 -c csi-s3
```

### Issues creating containers

1. Ensure feature gate `MountPropagation` is not set to `false`
2. Check the logs of the s3-driver:

```bash
kubectl logs -l app=csi-s3 -c csi-s3
```

## Development

This project can be built like any other go application.

```bash
go get -u github.com/ctrox/csi-s3
```

### Build executable

```bash
make build
```

### Tests

Currently the driver is tested by the [CSI Sanity Tester](https://github.com/kubernetes-csi/csi-test/tree/master/pkg/sanity). As end-to-end tests require S3 storage and a mounter like s3fs, this is best done in a docker container. A Dockerfile and the test script are in the `test` directory. The easiest way to run the tests is to just use the make command:

```bash
make test
```
