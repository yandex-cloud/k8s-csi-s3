/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yandex-cloud/k8s-csi-s3/pkg/endpoint"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"sync"
)

type Driver struct {
	endpoint string

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer

	driver *CSIDriver
	gcs    *groupControllerServer
}

var (
	vendorVersion = "v1.34.7"
	driverName    = "ru.yandex.s3.csi"
)

// New initializes the driver
func New(nodeID string, endpoint string) (*Driver, error) {
	d := NewCSIDriver(driverName, vendorVersion, nodeID)
	if d == nil {
		klog.Fatalln("Failed to initialize CSI Driver.")
	}

	s3Driver := &Driver{
		endpoint: endpoint,
		driver:   d,
	}
	return s3Driver, nil
}

func (s3 *Driver) newIdentityServer(d *CSIDriver) *identityServer {
	return &identityServer{d: d}
}

func (s3 *Driver) newControllerServer(d *CSIDriver) *controllerServer {
	return &controllerServer{d: d}
}

func (s3 *Driver) newNodeServer(d *CSIDriver) *nodeServer {
	return &nodeServer{d: d}
}

func NewNonBlockingGRPCServer() *NonBlockingGRPCServer {
	return &NonBlockingGRPCServer{}
}

type NonBlockingGRPCServer struct {
	wg      sync.WaitGroup
	server  *grpc.Server
	cleanup func()
}

func (s *NonBlockingGRPCServer) Start(endpoint string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer, gcs csi.GroupControllerServer) {
	s.wg.Add(1)
	go s.serve(endpoint, ids, cs, ns, gcs)
	return
}

func (s3 *Driver) Run() {
	klog.Infof("Driver: %v ", driverName)
	klog.Infof("Version: %v ", vendorVersion)
	// Initialize default library driver

	s3.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	s3.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER})
	s3.driver.AddNodeServiceCapabilities([]csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
		csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
	})

	// Create GRPC servers
	s3.ids = s3.newIdentityServer(s3.driver)
	s3.ns = s3.newNodeServer(s3.driver)
	s3.cs = s3.newControllerServer(s3.driver)
	s3.gcs = s3.newGroupControllerServer()

	s := NewNonBlockingGRPCServer()
	s.Start(s3.endpoint, s3.ids, s3.cs, s3.ns, s3.gcs)
	s.wg.Wait()
}

func (s *NonBlockingGRPCServer) serve(ep string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer, gcs csi.GroupControllerServer) {
	listener, cleanup, err := endpoint.Listen(ep)
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logGRPC),
	}
	server := grpc.NewServer(opts...)
	s.server = server
	s.cleanup = cleanup

	if ids != nil {
		csi.RegisterIdentityServer(server, ids)
	}
	if cs != nil {
		csi.RegisterControllerServer(server, cs)
	}
	if ns != nil {
		csi.RegisterNodeServer(server, ns)
	}
	if gcs != nil {
		csi.RegisterGroupControllerServer(server, gcs)
	}

	klog.Infof("Listening for connections on address: %#v", listener.Addr())
	server.Serve(listener)
}
