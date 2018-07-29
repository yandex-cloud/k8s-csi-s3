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

package main

import (
	"flag"
	"log"
	"os"

	"github.com/ctrox/csi-s3/pkg/s3"
)

func init() {
	flag.Set("logtostderr", "true")
}

var (
	endpoint        = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID          = flag.String("nodeid", "", "node id")
	accessKeyID     = flag.String("access-key-id", "", "S3 Access Key ID to use")
	secretAccessKey = flag.String("secret-access-key", "", "S3 Secret Access Key to use")
	s3endpoint      = flag.String("s3-endpoint", "", "S3 Endpoint URL to use")
	region          = flag.String("region", "", "S3 Region to use")
	mounter         = flag.String("mounter", "s3fs", "Specify which Mounter to use")
	encryptionKey   = flag.String("encryption-key", "", "Encryption key for file system (only used with s3ql)")
)

func main() {
	flag.Parse()

	cfg := &s3.Config{
		AccessKeyID:     *accessKeyID,
		SecretAccessKey: *secretAccessKey,
		Endpoint:        *s3endpoint,
		Region:          *region,
		Mounter:         *mounter,
		EncryptionKey:   *encryptionKey,
	}

	driver, err := s3.NewS3(*nodeID, *endpoint, cfg)
	if err != nil {
		log.Fatal(err)
	}
	driver.Run()
	os.Exit(0)
}
