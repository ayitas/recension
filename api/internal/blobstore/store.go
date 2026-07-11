package blobstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// DigestPrefix is the content-addressed digest scheme.
const DigestPrefix = "sha256:"

// Store is a content-addressed blob backend (MinIO / S3).
type Store struct {
	client *minio.Client
	bucket string
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// Digest computes sha256:<hex> for data.
func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return DigestPrefix + hex.EncodeToString(sum[:])
}

// ValidDigest reports whether dig looks like sha256:<64 hex>.
func ValidDigest(dig string) bool {
	if !strings.HasPrefix(dig, DigestPrefix) {
		return false
	}
	hexPart := strings.TrimPrefix(dig, DigestPrefix)
	if len(hexPart) != 64 {
		return false
	}
	_, err := hex.DecodeString(hexPart)
	return err == nil
}

func objectKey(digest string) string {
	hexPart := strings.TrimPrefix(digest, DigestPrefix)
	if len(hexPart) < 4 {
		return "blobs/" + digest
	}
	return "blobs/" + hexPart[:2] + "/" + hexPart[2:4] + "/" + hexPart
}

// New connects to MinIO/S3 and ensures the bucket exists.
func New(ctx context.Context, cfg Config) (*Store, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("blobstore: endpoint required")
	}
	if cfg.Bucket == "" {
		cfg.Bucket = "recension"
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("blobstore: connect: %w", err)
	}
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("blobstore: bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("blobstore: create bucket: %w", err)
		}
	}
	return &Store{client: client, bucket: cfg.Bucket}, nil
}

func (s *Store) Exists(ctx context.Context, digest string) (bool, error) {
	if !ValidDigest(digest) {
		return false, fmt.Errorf("invalid digest")
	}
	_, err := s.client.StatObject(ctx, s.bucket, objectKey(digest), minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Put stores raw bytes under digest. Digest must match the payload.
func (s *Store) Put(ctx context.Context, digest string, data []byte, mime string) error {
	if !ValidDigest(digest) {
		return fmt.Errorf("invalid digest")
	}
	if Digest(data) != digest {
		return fmt.Errorf("digest mismatch")
	}
	if mime == "" {
		mime = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, s.bucket, objectKey(digest),
		bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: mime},
	)
	return err
}

func (s *Store) Get(ctx context.Context, digest string) (io.ReadCloser, string, int64, error) {
	if !ValidDigest(digest) {
		return nil, "", 0, fmt.Errorf("invalid digest")
	}
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey(digest), minio.GetObjectOptions{})
	if err != nil {
		return nil, "", 0, err
	}
	st, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, "", 0, err
	}
	return obj, st.ContentType, st.Size, nil
}
