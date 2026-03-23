package media

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Tattsum/blog/backend/internal/application/upload"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Storage struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

func NewR2Storage(ctx context.Context, accountID, accessKeyID, secretAccessKey, bucket, publicBaseURL string) (*R2Storage, error) {
	accountID = strings.TrimSpace(accountID)
	accessKeyID = strings.TrimSpace(accessKeyID)
	secretAccessKey = strings.TrimSpace(secretAccessKey)
	bucket = strings.TrimSpace(bucket)
	publicBaseURL = strings.TrimSuffix(strings.TrimSpace(publicBaseURL), "/")
	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucket == "" || publicBaseURL == "" {
		return nil, fmt.Errorf("r2: accountID, accessKeyID, secretAccessKey, bucket, publicBaseURL are required")
	}
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
	})
	return &R2Storage{client: client, bucket: bucket, publicBaseURL: publicBaseURL}, nil
}

func (s *R2Storage) Put(ctx context.Context, key, contentType string, body io.Reader) (publicURL string, err error) {
	key = strings.TrimPrefix(key, "/")
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	return s.publicBaseURL + "/" + key, nil
}

var _ upload.MediaStorage = (*R2Storage)(nil)
