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
)

type driver struct {
	ids *IdentityServer
	cs  *ControllerServer
	name    string
	nodeID  string
	version string

	endpoint string

	ns *NodeServer
	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

var (
	vendorVersion = "v1.34.7"
	driverName    = "ru.yandex.s3.csi"
)

// New initializes the driver
func New(nodeID string, endpoint string) (*driver, error) {
	glog.Infof("Driver: %v version: %v", driverName, vendorVersion)

	d := &driver{
		name:     driverName,
		version:  vendorVersion,
		nodeID:   nodeID,
		endpoint: endpoint,
	}
	return d, nil
}

	
func (d *driver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) []*csi.VolumeCapability_AccessMode {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		glog.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, &csi.VolumeCapability_AccessMode{Mode: c})
	}
	d.cap = vca
	return vca
}

func (d *driver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability

	for _, c := range cl {
		glog.Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}

	d.cscap = csc

	return
}

func (s3 *driver) Run() {

	glog.Infof("Driver: %v ", driverName)
	glog.Infof("Version: %v ", vendorVersion)
	// Initialize default library driver

	s3.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	s3.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER})

	// Create GRPC servers
	s3.ids = NewDefaultIdentityServer(s3)
	s3.ns = NewNodeServer(s3)
	s3.cs = NewControllerServer(s3)
	s := NewNonBlockingGRPCServer()
	s.Start(s3.endpoint,
		s3.ids,
		s3.cs,
		s3.ns)
	s.Wait()
}
