package mounter

import (
	"fmt"
	"os"

	"github.com/ctrox/csi-s3/pkg/s3"
)

const (
	geesefsCmd    = "geesefs"
	defaultRegion = "us-east-1"
)

// Implements Mounter
type geesefsMounter struct {
	meta            *s3.FSMeta
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
}

func newGeeseFSMounter(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	region := cfg.Region
	// if endpoint is set we need a default region
	if region == "" && cfg.Endpoint != "" {
		region = defaultRegion
	}
	return &geesefsMounter{
		meta:            meta,
		endpoint:        cfg.Endpoint,
		region:          region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
	}, nil
}

func (geesefs *geesefsMounter) Stage(stageTarget string) error {
	return nil
}

func (geesefs *geesefsMounter) Unstage(stageTarget string) error {
	return nil
}

func (geesefs *geesefsMounter) Mount(source string, target string) error {
	fullPath := fmt.Sprintf("%s:%s", geesefs.meta.BucketName, geesefs.meta.Prefix)
	// FIXME Add memory limits
	args := []string{
		"--endpoint", geesefs.endpoint,
		"--region", geesefs.region,
		"-o", "allow_other",
		fullPath, target,
	}
	args = append(args, geesefs.meta.MountOptions...)
	os.Setenv("AWS_ACCESS_KEY_ID", geesefs.accessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", geesefs.secretAccessKey)
	return fuseMount(target, geesefsCmd, args)
}
