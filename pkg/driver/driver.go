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
	"github.com/golang/glog"

	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type driver struct {
	driver   *csicommon.CSIDriver
	endpoint string

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer
}

var (
	vendorVersion = "v1.34.7"
	driverName    = "ru.yandex.s3.csi"
)

// New initializes the driver
func New(nodeID string, endpoint string) (*driver, error) {
	d := csicommon.NewCSIDriver(driverName, vendorVersion, nodeID)
	if d == nil {
		glog.Fatalln("Failed to initialize CSI Driver.")
	}

	s3Driver := &driver{
		endpoint: endpoint,
		driver:   d,
	}
	return s3Driver, nil
}

func (s3 *driver) newIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}

func (s3 *driver) newControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
	}
}

func (s3 *driver) newNodeServer(d *csicommon.CSIDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
	}
}

func (s3 *driver) Run() {
	glog.Infof("Driver: %v ", driverName)
	glog.Infof("Version: %v ", vendorVersion)
	// Initialize default library driver

	s3.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	s3.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER})

	// Create GRPC servers
	s3.ids = s3.newIdentityServer(s3.driver)
	s3.ns = s3.newNodeServer(s3.driver)
	s3.cs = s3.newControllerServer(s3.driver)

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(s3.endpoint, s3.ids, s3.cs, s3.ns)
	s.Wait()
}
