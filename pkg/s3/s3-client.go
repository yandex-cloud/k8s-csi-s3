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
	fsPrefix     = "csi-fs"
)

type s3Client struct {
	cfg   *Config
	minio *minio.Client
	ctx   context.Context
}

type bucket struct {
	Name          string
	Mounter       string
	FSPath        string
	CapacityBytes int64
}

func newS3Client(cfg *Config) (*s3Client, error) {
	var client = &s3Client{}

	client.cfg = cfg
	u, err := url.Parse(client.cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	ssl := u.Scheme == "https"
	endpoint := u.Hostname()
	if u.Port() != "" {
		endpoint = u.Hostname() + ":" + u.Port()
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(client.cfg.AccessKeyID, client.cfg.SecretAccessKey, client.cfg.Region),
		Secure: ssl,
	})
	if err != nil {
		return nil, err
	}
	client.minio = minioClient
	client.ctx = context.Background()
	return client, nil
}

func newS3ClientFromSecrets(secrets map[string]string) (*s3Client, error) {
	return newS3Client(&Config{
		AccessKeyID:     secrets["accessKeyID"],
		SecretAccessKey: secrets["secretAccessKey"],
		Region:          secrets["region"],
		Endpoint:        secrets["endpoint"],
		// Mounter is set in the volume preferences, not secrets
		Mounter: "",
	})
}

func (client *s3Client) bucketExists(bucketName string) (bool, error) {
	return client.minio.BucketExists(client.ctx, bucketName)
}

func (client *s3Client) createBucket(bucketName string) error {
	return client.minio.MakeBucket(client.ctx, bucketName, minio.MakeBucketOptions{Region: client.cfg.Region})
}

func (client *s3Client) createPrefix(bucketName string, prefix string) error {
	_, err := client.minio.PutObject(client.ctx, bucketName, prefix+"/", bytes.NewReader([]byte("")), 0, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (client *s3Client) removeBucket(bucketName string) error {
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

	// ensure our prefix is also removed
	return client.minio.RemoveObject(client.ctx, bucketName, fsPrefix, minio.RemoveObjectOptions{})
}

func (client *s3Client) setBucket(bucket *bucket) error {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(bucket)
	opts := minio.PutObjectOptions{ContentType: "application/json"}
	_, err := client.minio.PutObject(client.ctx, bucket.Name, metadataName, b, int64(b.Len()), opts)
	return err
}

func (client *s3Client) getBucket(bucketName string) (*bucket, error) {
	opts := minio.GetObjectOptions{}
	obj, err := client.minio.GetObject(client.ctx, bucketName, metadataName, opts)
	if err != nil {
		return &bucket{}, err
	}
	objInfo, err := obj.Stat()
	if err != nil {
		return &bucket{}, err
	}
	b := make([]byte, objInfo.Size)
	_, err = obj.Read(b)

	if err != nil && err != io.EOF {
		return &bucket{}, err
	}
	var meta bucket
	err = json.Unmarshal(b, &meta)
	return &meta, err
}
