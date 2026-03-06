package mounter

import (
	"fmt"
	"path"

	"github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
)

// Implements Mounter
type rcloneMounter struct {
	meta                 *s3.FSMeta
	url                  string
	region               string
	accessKeyID          string
	secretAccessKey      string
	useIRSA              bool
	roleArn              string
	webIdentityTokenFile string
}

const (
	rcloneCmd = "rclone"
)

func newRcloneMounter(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	return &rcloneMounter{
		meta:                 meta,
		url:                  cfg.Endpoint,
		region:               cfg.Region,
		accessKeyID:          cfg.AccessKeyID,
		secretAccessKey:      cfg.SecretAccessKey,
		useIRSA:              cfg.UseIRSA,
		roleArn:              cfg.RoleArn,
		webIdentityTokenFile: cfg.WebIdentityTokenFile,
	}, nil
}

func (rclone *rcloneMounter) Mount(target, volumeID string) error {
	args := []string{
		"mount",
		fmt.Sprintf(":s3:%s", path.Join(rclone.meta.BucketName, rclone.meta.Prefix)),
		target,
		"--daemon",
		"--s3-provider=AWS",
		"--s3-env-auth=true",
		fmt.Sprintf("--s3-endpoint=%s", rclone.url),
		"--allow-other",
		"--vfs-cache-mode=writes",
	}
	if rclone.region != "" {
		args = append(args, fmt.Sprintf("--s3-region=%s", rclone.region))
	}
	args = append(args, rclone.meta.MountOptions...)
	var envs []string
	if rclone.useIRSA {
		if rclone.roleArn != "" {
			envs = append(envs, "AWS_ROLE_ARN="+rclone.roleArn)
		}
		if rclone.webIdentityTokenFile != "" {
			envs = append(envs, "AWS_WEB_IDENTITY_TOKEN_FILE="+rclone.webIdentityTokenFile)
		}
	} else {
		envs = []string{
			"AWS_ACCESS_KEY_ID=" + rclone.accessKeyID,
			"AWS_SECRET_ACCESS_KEY=" + rclone.secretAccessKey,
		}
	}
	return fuseMount(target, rcloneCmd, args, envs)
}
