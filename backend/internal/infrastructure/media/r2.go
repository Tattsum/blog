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

// R2Storage は Cloudflare R2（S3 互換 API）に保存し、公開 URL のベースを付与して返す。
// パブリックアクセスは R2 ダッシュボードで r2.dev サブドメインまたはカスタムドメインを有効にすること。
type R2Storage struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

// NewR2Storage は accountID と R2 API トークンで S3 互換クライアントを生成する。
// publicBaseURL はアップロード後の公開 URL のベース（例: https://pub-xxxx.r2.dev または https://media.example.com）。末尾スラッシュなし。
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

// Put は body を key で R2 バケットに書き込む。公開 URL は publicBaseURL + "/" + key で返す。
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
