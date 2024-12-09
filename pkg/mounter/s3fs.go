package mounter

import (
        "fmt"
        "os"

        "github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
        "github.com/google/uuid"
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

func (s3fs *s3fsMounter) Mount(target, volumeID string) error {
    pwFileName, err := writes3fsPass(s3fs.pwFileContent)
    if err != nil {
        return err
    }
    args := []string{
        fmt.Sprintf("%s:/%s", s3fs.meta.BucketName, s3fs.meta.Prefix),
        target,
        "-o", "use_path_request_style",
        "-o", fmt.Sprintf("url=%s", s3fs.url),
        "-o", "allow_other",
        "-o", "mp_umask=000",
        "-o", fmt.Sprintf("passwd_file=%s", pwFileName),
    }
    if s3fs.region != "" {
        args = append(args, "-o", fmt.Sprintf("endpoint=%s", s3fs.region))
    }
    args = append(args, s3fs.meta.MountOptions...)
    return fuseMount(target, s3fsCmd, args, nil)
}

func writes3fsPass(pwFileContent string) (string, error) {
    tempDir := os.TempDir()
    uuid := uuid.New()
    pwFileName := fmt.Sprintf("%s/%s", tempDir, uuid)
    pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
    if err != nil {
        return "OpenFile Failed", err
    }
    _, err = pwFile.WriteString(pwFileContent)
    if err != nil {
        return "WriteString failed", err
    }
    pwFile.Close()
    return pwFileName, nil
}