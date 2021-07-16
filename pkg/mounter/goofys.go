package mounter

import (
	"fmt"
	"os"
	"path"
	"strings"

	"context"

	"github.com/ctrox/csi-s3/pkg/s3"
	goofysApi "github.com/kahing/goofys/api"
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
	mountOptions := make(map[string]string)
	for _, opt := range goofys.meta.MountOptions {
		s := 0
		if len(opt) >= 2 && opt[0] == '-' && opt[1] == byte('-') {
			s = 2
		}
		p := strings.Index(opt, "=")
		if p > 0 {
			mountOptions[string(opt[s : p])] = string(opt[p : ])
		} else {
			mountOptions[string(opt[s : ])] = ""
		}
	}
	mountOptions["allow_other"] = ""
	goofysCfg := &goofysApi.Config{
		MountPoint: target,
		Endpoint:   goofys.endpoint,
		Region:     goofys.region,
		DirMode:    0755,
		FileMode:   0644,
		MountOptions: mountOptions,
	}

	os.Setenv("AWS_ACCESS_KEY_ID", goofys.accessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", goofys.secretAccessKey)
	fullPath := fmt.Sprintf("%s:%s", goofys.meta.BucketName, path.Join(goofys.meta.Prefix, goofys.meta.FSPath))

	_, _, err := goofysApi.Mount(context.Background(), fullPath, goofysCfg)

	if err != nil {
		return fmt.Errorf("Error mounting via goofys: %s", err)
	}
	return nil
}
