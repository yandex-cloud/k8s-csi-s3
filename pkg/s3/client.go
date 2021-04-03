package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

type Bucket struct {
	Name          string
	Mounter       string
	FSPath        string
	CapacityBytes int64
	CreatedByCsi  bool
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
	_, err := client.minio.PutObject(client.ctx, bucketName, prefix+"/", bytes.NewReader([]byte("")), 0, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (client *s3Client) RemovePrefix(bucketName string, prefix string) error {
	return client.minio.RemoveObject(client.ctx, bucketName, prefix, minio.RemoveObjectOptions{})
}

func (client *s3Client) RemoveBucket(bucketName string) error {
	if err := client.emptyBucket(bucketName); err != nil {
		return err
	}
	return client.minio.RemoveBucket(client.ctx, bucketName)
}

func (client *s3Client) emptyBucket(bucketName string) error {
	objectsCh := make(chan minio.ObjectInfo)
	var listErr error

	go func() {
		defer close(objectsCh)

		doneCh := make(chan struct{})

		defer close(doneCh)

		for object := range client.minio.ListObjects(
			client.ctx,
			bucketName,
			minio.ListObjectsOptions{Prefix: "", Recursive: true}) {
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
		for e := range errorCh {
			glog.Errorf("Failed to remove object %s, error: %s", e.ObjectName, e.Err)
		}
		if len(errorCh) != 0 {
			return fmt.Errorf("Failed to remove all objects of bucket %s", bucketName)
		}
	}

	return nil
}

func (client *s3Client) SetBucket(bucket *Bucket) error {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(bucket)
	opts := minio.PutObjectOptions{ContentType: "application/json"}
	_, err := client.minio.PutObject(client.ctx, bucket.Name, metadataName, b, int64(b.Len()), opts)
	return err
}

func (client *s3Client) GetBucket(bucketName string) (*Bucket, error) {
	opts := minio.GetObjectOptions{}
	obj, err := client.minio.GetObject(client.ctx, bucketName, metadataName, opts)
	if err != nil {
		return &Bucket{}, err
	}
	objInfo, err := obj.Stat()
	if err != nil {
		return &Bucket{}, err
	}
	b := make([]byte, objInfo.Size)
	_, err = obj.Read(b)

	if err != nil && err != io.EOF {
		return &Bucket{}, err
	}
	var meta Bucket
	err = json.Unmarshal(b, &meta)
	return &meta, err
}
