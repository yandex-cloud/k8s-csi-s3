package mounter

import (
	"fmt"
	"os"

	"github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
)

// Implements Mounter
type s3fsMounter struct {
	meta                 *s3.FSMeta
	url                  string
	region               string
	pwFileContent        string
	useIRSA              bool
	roleArn              string
	webIdentityTokenFile string
}

const (
	s3fsCmd = "s3fs"
)

func newS3fsMounter(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	return &s3fsMounter{
		meta:                 meta,
		url:                  cfg.Endpoint,
		region:               cfg.Region,
		pwFileContent:        cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
		useIRSA:              cfg.UseIRSA,
		roleArn:              cfg.RoleArn,
		webIdentityTokenFile: cfg.WebIdentityTokenFile,
	}, nil
}

func (s3fs *s3fsMounter) Mount(target, volumeID string) error {
	var envs []string
	if s3fs.useIRSA {
		// s3fs supports IAM role auth; skip password file
		if s3fs.roleArn != "" {
			envs = append(envs, "AWS_ROLE_ARN="+s3fs.roleArn)
		}
		if s3fs.webIdentityTokenFile != "" {
			envs = append(envs, "AWS_WEB_IDENTITY_TOKEN_FILE="+s3fs.webIdentityTokenFile)
		}
	} else {
		if err := writes3fsPass(s3fs.pwFileContent); err != nil {
			return err
		}
	}
	args := []string{
		fmt.Sprintf("%s:/%s", s3fs.meta.BucketName, s3fs.meta.Prefix),
		target,
		"-o", "use_path_request_style",
		"-o", fmt.Sprintf("url=%s", s3fs.url),
		"-o", "allow_other",
		"-o", "mp_umask=000",
	}
	if s3fs.useIRSA {
		args = append(args, "-o", "iam_role=auto")
	}
	if s3fs.region != "" {
		args = append(args, "-o", fmt.Sprintf("endpoint=%s", s3fs.region))
	}
	args = append(args, s3fs.meta.MountOptions...)
	return fuseMount(target, s3fsCmd, args, envs)
}

func writes3fsPass(pwFileContent string) error {
	pwFileName := fmt.Sprintf("%s/.passwd-s3fs", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(pwFileContent)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}
