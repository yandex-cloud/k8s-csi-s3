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
	"context"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"google.golang.org/grpc"
)

type driver struct {
	name     string
	version  string
	nodeID   string
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
	d := &driver{
		name:     driverName,
		version:  vendorVersion,
		nodeID:   nodeID,
		endpoint: endpoint,
	}
	return d, nil
}

func (d *driver) Run() {
	glog.Infof("Driver: %v ", d.name)
	glog.Infof("Version: %v ", d.version)

	d.ids = &identityServer{driver: d}
	d.ns = &nodeServer{driver: d}
	d.cs = &controllerServer{driver: d}

	// Parse endpoint
	u, err := url.Parse(d.endpoint)
	if err != nil {
		glog.Fatalf("Failed to parse endpoint %s: %v", d.endpoint, err)
	}

	var addr string
	switch u.Scheme {
	case "unix":
		addr = u.Path
		if err := os.MkdirAll(path.Dir(addr), 0750); err != nil {
			glog.Fatalf("Failed to create directory for socket: %v", err)
		}
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			glog.Fatalf("Failed to remove existing socket: %v", err)
		}
	case "tcp":
		addr = u.Host
	default:
		glog.Fatalf("Unsupported protocol: %s", u.Scheme)
	}

	listener, err := net.Listen(u.Scheme, addr)
	if err != nil {
		glog.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	// Ensure socket file is cleaned up on exit
	if u.Scheme == "unix" {
		absAddr, err := filepath.Abs(addr)
		if err == nil {
			defer os.Remove(absAddr)
		}
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logGRPC),
	}
	server := grpc.NewServer(opts...)

	csi.RegisterIdentityServer(server, d.ids)
	csi.RegisterControllerServer(server, d.cs)
	csi.RegisterNodeServer(server, d.ns)

	glog.Infof("Listening for connections on address: %#v", listener.Addr())
	if err := server.Serve(listener); err != nil {
		glog.Fatalf("Failed to serve: %v", err)
	}
}

// logGRPC logs all gRPC calls
func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	glog.V(3).Infof("GRPC call: %s", info.FullMethod)
	glog.V(5).Infof("GRPC request: %+v", req)
	resp, err := handler(ctx, req)
	if err != nil {
		glog.Errorf("GRPC error: %v", err)
	} else {
		glog.V(5).Infof("GRPC response: %+v", resp)
	}
	return resp, err
}
