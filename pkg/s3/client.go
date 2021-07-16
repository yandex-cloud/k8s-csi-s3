package s3

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/golang/glog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	metadataName = ".metadata.json"
)

type s3Client struct {
	Config *Config
	minio  *minio.Client
	ctx    context.Context
}

// Config holds values to configure the driver
type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Endpoint        string
	Mounter         string
}

type FSMeta struct {
	BucketName    string `json:"Name"`
	Prefix        string `json:"Prefix"`
	Mounter       string `json:"Mounter"`
	MountOptions  []string `json:"MountOptions"`
	CapacityBytes int64  `json:"CapacityBytes"`
}

func NewClient(cfg *Config) (*s3Client, error) {
	var client = &s3Client{}

	client.Config = cfg
	u, err := url.Parse(client.Config.Endpoint)
	if err != nil {
		return nil, err
	}
	ssl := u.Scheme == "https"
	endpoint := u.Hostname()
	if u.Port() != "" {
		endpoint = u.Hostname() + ":" + u.Port()
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(client.Config.AccessKeyID, client.Config.SecretAccessKey, client.Config.Region),
		Secure: ssl,
	})
	if err != nil {
		return nil, err
	}
	client.minio = minioClient
	client.ctx = context.Background()
	return client, nil
}

func NewClientFromSecret(secret map[string]string) (*s3Client, error) {
	return NewClient(&Config{
		AccessKeyID:     secret["accessKeyID"],
		SecretAccessKey: secret["secretAccessKey"],
		Region:          secret["region"],
		Endpoint:        secret["endpoint"],
		// Mounter is set in the volume preferences, not secrets
		Mounter: "",
	})
}

func (client *s3Client) BucketExists(bucketName string) (bool, error) {
	return client.minio.BucketExists(client.ctx, bucketName)
}

func (client *s3Client) CreateBucket(bucketName string) error {
	return client.minio.MakeBucket(client.ctx, bucketName, minio.MakeBucketOptions{Region: client.Config.Region})
}

func (client *s3Client) CreatePrefix(bucketName string, prefix string) error {
	if prefix != "" {
		_, err := client.minio.PutObject(client.ctx, bucketName, prefix+"/", bytes.NewReader([]byte("")), 0, minio.PutObjectOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *s3Client) RemovePrefix(bucketName string, prefix string) error {
	var err error

	if err = client.removeObjects(bucketName, prefix); err == nil {
		return client.minio.RemoveObject(client.ctx, bucketName, prefix, minio.RemoveObjectOptions{})
	}

	glog.Warningf("removeObjects failed with: %s, will try removeObjectsOneByOne", err)

	if err = client.removeObjectsOneByOne(bucketName, prefix); err == nil {
		return client.minio.RemoveObject(client.ctx, bucketName, prefix, minio.RemoveObjectOptions{})
	}

	return err
}

func (client *s3Client) RemoveBucket(bucketName string) error {
	var err error

	if err = client.removeObjects(bucketName, ""); err == nil {
		return client.minio.RemoveBucket(client.ctx, bucketName)
	}

	glog.Warningf("removeObjects failed with: %s, will try removeObjectsOneByOne", err)

	if err = client.removeObjectsOneByOne(bucketName, ""); err == nil {
		return client.minio.RemoveBucket(client.ctx, bucketName)
	}

	return err
}

func (client *s3Client) removeObjects(bucketName, prefix string) error {
	objectsCh := make(chan minio.ObjectInfo)
	var listErr error

	go func() {
		defer close(objectsCh)

		for object := range client.minio.ListObjects(
			client.ctx,
			bucketName,
			minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			if object.Err != nil {
				listErr = object.Err
				return
			}
			objectsCh <- object
		}
	}()

	if listErr != nil {
		glog.Error("Error listing objects", listErr)
		return listErr
	}

	select {
	default:
		opts := minio.RemoveObjectsOptions{
			GovernanceBypass: true,
		}
		errorCh := client.minio.RemoveObjects(client.ctx, bucketName, objectsCh, opts)
		haveErrWhenRemoveObjects := false
		for e := range errorCh {
			glog.Errorf("Failed to remove object %s, error: %s", e.ObjectName, e.Err)
			haveErrWhenRemoveObjects = true
		}
		if haveErrWhenRemoveObjects {
			return fmt.Errorf("Failed to remove all objects of bucket %s", bucketName)
		}
	}

	return nil
}

// will delete files one by one without file lock
func (client *s3Client) removeObjectsOneByOne(bucketName, prefix string) error {
	objectsCh := make(chan minio.ObjectInfo, 1)
	removeErrCh := make(chan minio.RemoveObjectError, 1)
	var listErr error

	go func() {
		defer close(objectsCh)

		for object := range client.minio.ListObjects(client.ctx, bucketName,
			minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			if object.Err != nil {
				listErr = object.Err
				return
			}
			objectsCh <- object
		}
	}()

	if listErr != nil {
		glog.Error("Error listing objects", listErr)
		return listErr
	}

	go func() {
		defer close(removeErrCh)

		for object := range objectsCh {
			err := client.minio.RemoveObject(client.ctx, bucketName, object.Key,
				minio.RemoveObjectOptions{VersionID: object.VersionID})
			if err != nil {
				removeErrCh <- minio.RemoveObjectError{
					ObjectName: object.Key,
					VersionID:  object.VersionID,
					Err:        err,
				}
			}
		}
	}()

	haveErrWhenRemoveObjects := false
	for e := range removeErrCh {
		glog.Errorf("Failed to remove object %s, error: %s", e.ObjectName, e.Err)
		haveErrWhenRemoveObjects = true
	}
	if haveErrWhenRemoveObjects {
		return fmt.Errorf("Failed to remove all objects of path %s", bucketName)
	}

	return nil
}
