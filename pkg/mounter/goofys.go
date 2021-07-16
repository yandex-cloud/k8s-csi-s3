package mounter

import (
	"fmt"
	"os"

	"github.com/ctrox/csi-s3/pkg/s3"
)

const (
	goofysCmd     = "goofys"
	defaultRegion = "us-east-1"
)

// Implements Mounter
type goofysMounter struct {
	meta            *s3.FSMeta
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
}

func newGoofysMounter(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	region := cfg.Region
	// if endpoint is set we need a default region
	if region == "" && cfg.Endpoint != "" {
		region = defaultRegion
	}
	return &goofysMounter{
		meta:            meta,
		endpoint:        cfg.Endpoint,
		region:          region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
	}, nil
}

func (goofys *goofysMounter) Stage(stageTarget string) error {
	return nil
}

func (goofys *goofysMounter) Unstage(stageTarget string) error {
	return nil
}

func (goofys *goofysMounter) Mount(source string, target string) error {
	fullPath := fmt.Sprintf("%s:%s", goofys.meta.BucketName, goofys.meta.Prefix)
	args := []string{
		"--endpoint", goofys.endpoint,
		"--region", goofys.region,
		"-o", "allow_other",
		fullPath, target,
	}
	args = append(args, goofys.meta.MountOptions...)
	os.Setenv("AWS_ACCESS_KEY_ID", goofys.accessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", goofys.secretAccessKey)
	return fuseMount(target, goofysCmd, args)
}
