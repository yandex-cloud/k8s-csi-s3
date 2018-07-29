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

package s3

import (
	"fmt"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
	*s3
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	volumeID := req.GetName()

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid create volume req: %v", req)
		return nil, err
	}

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}

	capacityBytes := int64(req.GetCapacityRange().GetRequiredBytes())
	params := req.GetParameters()
	mounter := params[mounterTypeKey]

	glog.V(4).Infof("Got a request to create volume %s", volumeID)

	exists, err := cs.s3.client.bucketExists(volumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %v", err)
	}
	if exists {
		var b *bucket
		b, err = cs.s3.client.getBucket(volumeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get bucket metadata: %v", err)
		}
		// Check if volume capacity requested is bigger than the already existing capacity
		if capacityBytes > b.CapacityBytes {
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Volume with the same name: %s but with smaller size already exist", volumeID))
		}
	} else {
		if err = cs.s3.client.createBucket(volumeID); err != nil {
			return nil, fmt.Errorf("failed to create volume: %v", err)
		}
		if err = cs.s3.client.createPrefix(volumeID, fsPrefix); err != nil {
			return nil, fmt.Errorf("failed to create prefix: %v", err)
		}
	}
	b := &bucket{
		Name:          volumeID,
		Mounter:       mounter,
		CapacityBytes: capacityBytes,
		FSPath:        fsPrefix,
	}
	if err := cs.s3.client.setBucket(b); err != nil {
		return nil, fmt.Errorf("Error setting bucket metadata: %v", err)
	}

	glog.V(4).Infof("create volume %s", volumeID)
	s3Vol := s3Volume{}
	s3Vol.VolName = volumeID
	s3Vol.VolID = volumeID
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id:            volumeID,
			CapacityBytes: capacityBytes,
			Attributes:    req.GetParameters(),
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volumeID := req.GetVolumeId()

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("Invalid delete volume req: %v", req)
		return nil, err
	}
	glog.V(4).Infof("Deleting volume %s", volumeID)

	exists, err := cs.s3.client.bucketExists(volumeID)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := cs.s3.client.removeBucket(volumeID); err != nil {
			glog.V(3).Infof("Failed to remove volume: %v", err)
			return nil, err
		}
	} else {
		glog.V(5).Infof("Bucket %s does not exist, ignoring request", volumeID)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities missing in request")
	}

	exists, err := cs.s3.client.bucketExists(req.GetVolumeId())
	if err != nil {
		return nil, err
	}
	if !exists {
		// return an error if the volume requested does not exist
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Volume with id %s does not exist", req.GetVolumeId()))
	}

	for _, cap := range req.VolumeCapabilities {
		if cap.GetAccessMode().GetMode() != csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
			return &csi.ValidateVolumeCapabilitiesResponse{Supported: false, Message: ""}, nil
		}
	}
	return &csi.ValidateVolumeCapabilitiesResponse{Supported: true, Message: ""}, nil
}
