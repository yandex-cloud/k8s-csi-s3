package mounter

import (
	"fmt"
	"os"
	"path"

	"github.com/ctrox/csi-s3/pkg/s3"
)

// Implements Mounter
type s3fsMounter struct {
	meta          *s3.FSMeta
	url           string
	region        string
	pwFileContent string
}

const (
	s3fsCmd = "s3fs"
)

func newS3fsMounter(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	return &s3fsMounter{
		meta:          meta,
		url:           cfg.Endpoint,
		region:        cfg.Region,
		pwFileContent: cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

func (s3fs *s3fsMounter) Stage(stageTarget string) error {
	return nil
}

func (s3fs *s3fsMounter) Unstage(stageTarget string) error {
	return nil
}

func (s3fs *s3fsMounter) Mount(source string, target string) error {
	if err := writes3fsPass(s3fs.pwFileContent); err != nil {
		return err
	}
	args := []string{
		fmt.Sprintf("%s:/%s", s3fs.meta.BucketName, path.Join(s3fs.meta.Prefix, s3fs.meta.FSPath)),
		target,
		"-o", "use_path_request_style",
		"-o", fmt.Sprintf("url=%s", s3fs.url),
		"-o", fmt.Sprintf("endpoint=%s", s3fs.region),
		"-o", "allow_other",
		"-o", "mp_umask=000",
	}
	return fuseMount(target, s3fsCmd, args)
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
