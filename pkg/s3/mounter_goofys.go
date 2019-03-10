package s3

import (
	"fmt"
	"os"

	"context"

	goofysApi "github.com/kahing/goofys/api"
)

const (
	goofysCmd     = "goofys"
	defaultRegion = "us-east-1"
)

// Implements Mounter
type goofysMounter struct {
	bucket          *bucket
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
}

func newGoofysMounter(b *bucket, cfg *Config) (Mounter, error) {
	region := cfg.Region
	// if endpoint is set we need a default region
	if region == "" && cfg.Endpoint != "" {
		region = defaultRegion
	}
	return &goofysMounter{
		bucket:          b,
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
	goofysCfg := &goofysApi.Config{
		MountPoint: target,
		Endpoint:   goofys.endpoint,
		Region:     goofys.region,
		DirMode:    0755,
		FileMode:   0644,
		MountOptions: map[string]string{
			"allow_other": "",
		},
	}

	os.Setenv("AWS_ACCESS_KEY_ID", goofys.accessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", goofys.secretAccessKey)
	fullPath := fmt.Sprintf("%s:%s", goofys.bucket.Name, goofys.bucket.FSPath)

	_, _, err := goofysApi.Mount(context.Background(), fullPath, goofysCfg)

	if err != nil {
		return fmt.Errorf("Error mounting via goofys: %s", err)
	}
	return nil
}
