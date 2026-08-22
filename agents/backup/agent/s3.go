package agent

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func CopyS3(ctx context.Context, src Source, access, secret string, dest *ObjectStore, destKeyPrefix string, onObj func(key string, n int64)) (BackupResult, error) {
	var res BackupResult
	res.Engine = "s3copy"
	res.Kind = KindFull
	if src.S3Bucket == "" {
		return res, fmt.Errorf("s3 bucket is empty")
	}
	endpoint := strings.TrimSpace(src.Host)
	if endpoint == "" {
		return res, fmt.Errorf("s3 endpoint (host) is empty")
	}
	cl, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, secret, ""),
		Secure: src.S3Secure,
		Region: src.S3Region,
	})
	if err != nil {
		return res, err
	}
	prefix := strings.TrimLeft(src.S3Prefix, "/")
	for obj := range cl.ListObjects(ctx, src.S3Bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil {
			return res, obj.Err
		}
		if strings.HasSuffix(obj.Key, "/") {
			continue
		}
		rc, err := cl.GetObject(ctx, src.S3Bucket, obj.Key, minio.GetObjectOptions{})
		if err != nil {
			return res, err
		}
		rel := obj.Key
		if prefix != "" {
			rel = strings.TrimPrefix(obj.Key, prefix)
			rel = strings.TrimPrefix(rel, "/")
		}
		key := path.Join(destKeyPrefix, rel)
		n, err := dest.Put(ctx, key, rc, obj.Size, obj.ContentType)
		rc.Close()
		if err != nil {
			return res, err
		}
		res.Files++
		res.BytesRead += obj.Size
		res.BytesWritten += n
		if onObj != nil {
			onObj(obj.Key, n)
		}
	}
	res.Bucket = dest.Bucket()
	res.Key = destKeyPrefix + "/"
	if res.Files == 0 {
		return res, fmt.Errorf("no objects under s3://%s/%s", src.S3Bucket, prefix)
	}
	return res, nil
}

func RestoreS3Copy(ctx context.Context, dest *ObjectStore, srcPrefix, destEndpoint, destBucket, destAccess, destSecret, destPrefix string, secure bool, region string) (int, error) {
	if destEndpoint == "" || destBucket == "" {
		return 0, fmt.Errorf("restore s3 destination endpoint and bucket are required")
	}
	cl, err := minio.New(destEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(destAccess, destSecret, ""),
		Secure: secure,
		Region: region,
	})
	if err != nil {
		return 0, err
	}
	keys, err := dest.List(ctx, srcPrefix)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, key := range keys {
		rc, err := dest.Get(ctx, key)
		if err != nil {
			return n, err
		}
		rel := strings.TrimPrefix(key, strings.TrimSuffix(srcPrefix, "/")+"/")
		out := path.Join(destPrefix, rel)
		_, err = cl.PutObject(ctx, destBucket, out, rc, -1, minio.PutObjectOptions{})
		rc.Close()
		if err != nil {
			return n, err
		}
		n++
	}
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}
