package driver

import "github.com/container-storage-interface/spec/lib/go/csi"

type groupControllerServer struct {
	csi.UnimplementedGroupControllerServer
}

func (s3 *Driver) newGroupControllerServer() *groupControllerServer {
	return &groupControllerServer{}
}
