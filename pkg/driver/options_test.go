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
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yandex-cloud/k8s-csi-s3/pkg/mounter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestParsedGeeseFSOptionsValidation(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		want         []string
		wantRejected bool
	}{
		{
			name:         "quoted flag and value",
			input:        `"--setgid" "65534"`,
			want:         []string{"--setgid", "65534"},
			wantRejected: true,
		},
		{
			name:         "quoted part of flag",
			input:        `--"setuid"=0`,
			want:         []string{"--setuid=0"},
			wantRejected: true,
		},
		{
			name:  "allowed options keep order and values",
			input: `--memory-limit=4000 --dir-mode "0755"`,
			want:  []string{"--memory-limit=4000", "--dir-mode", "0755"},
		},
		{
			name:  "escaped quote in value",
			input: `--refresh-attr "name\"with-quote"`,
			want:  []string{"--refresh-attr", `name"with-quote`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := getMeta("", "", map[string]string{
				mounter.OptionsKey: tt.input,
			})

			if len(meta.MountOptions) != len(tt.want) {
				t.Fatalf("parsed %#v, want %#v", meta.MountOptions, tt.want)
			}
			for i := range tt.want {
				if meta.MountOptions[i] != tt.want[i] {
					t.Fatalf("argument %d = %q, want %q", i, meta.MountOptions[i], tt.want[i])
				}
			}

			err := mounter.ValidateGeeseFSOptions(meta.MountOptions)
			if tt.wantRejected && err == nil {
				t.Fatalf("forbidden option was accepted: %q", tt.input)
			}
			if !tt.wantRejected && err != nil {
				t.Fatalf("unexpected rejection: %v", err)
			}
		})
	}
}

func TestCreateVolumeRejectsGeeseFSOptionsBeforeS3(t *testing.T) {
	req := &csi.CreateVolumeRequest{
		Name: "test-volume",
		Parameters: map[string]string{
			mounter.OptionsKey: `"--setgid" "65534"`,
		},
		VolumeCapabilities: []*csi.VolumeCapability{{}},
		// Secrets are intentionally omitted: validation must reject the request before accessing S3.
	}

	_, err := (&controllerServer{}).CreateVolume(context.Background(), req)
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Fatalf("CreateVolume status = %v, want InvalidArgument; error: %v", got, err)
	}
}

func TestCreateVolumeDoesNotValidateOtherMountersAsGeeseFS(t *testing.T) {
	for _, mounterType := range []string{"s3fs", "rclone"} {
		t.Run(mounterType, func(t *testing.T) {
			req := &csi.CreateVolumeRequest{
				Name: "test-volume",
				Parameters: map[string]string{
					mounter.TypeKey:    mounterType,
					mounter.OptionsKey: "--setuid=0",
				},
				VolumeCapabilities: []*csi.VolumeCapability{{}},
				Secrets: map[string]string{
					"endpoint": "://invalid",
				},
			}

			_, err := (&controllerServer{}).CreateVolume(context.Background(), req)
			if err == nil || status.Code(err) == codes.InvalidArgument {
				t.Fatalf("CreateVolume error = %v; expected S3 configuration error, not GeeseFS validation", err)
			}
		})
	}
}
