package s3

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/util/mount"
)

// Implements Mounter
type s3backerMounter struct {
	bucket          *bucket
	url             string
	region          string
	accessKeyID     string
	secretAccessKey string
	ssl             bool
}

const (
	s3backerCmd    = "s3backer"
	s3backerFsType = "xfs"
	s3backerDevice = "file"
	// blockSize to use in k
	s3backerBlockSize   = "128k"
	s3backerDefaultSize = 1024 * 1024 * 1024 // 1GiB
	// S3backerLoopDevice the loop device required by s3backer
	S3backerLoopDevice = "/dev/loop0"
)

func newS3backerMounter(bucket *bucket, cfg *Config) (Mounter, error) {
	url, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, bucket.Name, bucket.FSPath)
	// s3backer cannot work with 0 size volumes
	if bucket.CapacityBytes == 0 {
		bucket.CapacityBytes = s3backerDefaultSize
	}
	s3backer := &s3backerMounter{
		bucket:          bucket,
		url:             cfg.Endpoint,
		region:          cfg.Region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
		ssl:             url.Scheme == "https",
	}

	return s3backer, s3backer.writePasswd()
}

func (s3backer *s3backerMounter) String() string {
	return s3backer.bucket.Name
}

func (s3backer *s3backerMounter) Stage(stageTarget string) error {
	// s3backer uses the loop device
	if err := createLoopDevice(S3backerLoopDevice); err != nil {
		return err
	}
	// s3backer requires two mounts
	// first mount will fuse mount the bucket to a single 'file'
	if err := s3backer.mountInit(stageTarget); err != nil {
		return err
	}
	// ensure 'file' device is formatted
	err := formatFs(s3backerFsType, path.Join(stageTarget, s3backerDevice))
	if err != nil {
		fuseUnmount(stageTarget)
	}
	return err
}

func (s3backer *s3backerMounter) Unstage(stageTarget string) error {
	// Unmount the s3backer fuse mount
	return fuseUnmount(stageTarget)
}

func (s3backer *s3backerMounter) Mount(source string, target string) error {
	device := path.Join(source, s3backerDevice)
	// second mount will mount the 'file' as a filesystem
	err := mount.New("").Mount(device, target, s3backerFsType, []string{})
	if err != nil {
		// cleanup fuse mount
		fuseUnmount(target)
		return err
	}
	return nil
}

func (s3backer *s3backerMounter) mountInit(path string) error {
	args := []string{
		fmt.Sprintf("--blockSize=%s", s3backerBlockSize),
		fmt.Sprintf("--size=%v", s3backer.bucket.CapacityBytes),
		fmt.Sprintf("--prefix=%s/", s3backer.bucket.FSPath),
		"--listBlocks",
		s3backer.bucket.Name,
		path,
	}
	if s3backer.region != "" {
		args = append(args, fmt.Sprintf("--region=%s", s3backer.region))
	} else {
		// only set baseURL if not on AWS (region is not set)
		// baseURL must end with /
		args = append(args, fmt.Sprintf("--baseURL=%s/", s3backer.url))
	}
	if s3backer.ssl {
		args = append(args, "--ssl")
	}

	return fuseMount(path, s3backerCmd, args)
}

func (s3backer *s3backerMounter) writePasswd() error {
	pwFileName := fmt.Sprintf("%s/.s3backer_passwd", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(s3backer.accessKeyID + ":" + s3backer.secretAccessKey)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}

func formatFs(fsType string, device string) error {
	diskMounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}
	format, err := diskMounter.GetDiskFormat(device)
	if err != nil {
		return err
	}
	if format != "" {
		glog.Infof("Disk %s is already formatted with format %s", device, format)
		return nil
	}
	args := []string{
		device,
	}
	cmd := exec.Command("mkfs."+fsType, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error formatting disk: %s", out)
	}
	glog.Infof("Formatting fs with type %s", fsType)
	return nil
}
