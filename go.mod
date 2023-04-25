module github.com/yandex-cloud/k8s-csi-s3

go 1.15

require (
	github.com/container-storage-interface/spec v1.8.0
	github.com/coreos/go-systemd/v22 v22.5.0
	github.com/godbus/dbus/v5 v5.0.4
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/kubernetes-csi/csi-lib-utils v0.6.1 // indirect
	github.com/kubernetes-csi/csi-test v2.0.0+incompatible
	github.com/kubernetes-csi/drivers v1.0.2
	github.com/minio/minio-go/v7 v7.0.5
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.27.4
	golang.org/x/net v0.8.0
	google.golang.org/grpc v1.31.0
	k8s.io/api v0.27.1
	k8s.io/apimachinery v0.27.1
	k8s.io/client-go v0.27.1
	k8s.io/klog v0.2.0 // indirect
	k8s.io/kubernetes v1.13.4
)
