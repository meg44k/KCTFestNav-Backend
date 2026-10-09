package storage

import (
	"bytes"
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// 本番の置き場所。Cloudflare R2(S3 互換)。来場者には baseURL(img.kctfes.app)から配る
type R2 struct {
	client  *s3.Client
	bucket  string
	baseURL string
}

func NewR2(accountID, bucket, keyID, secret, baseURL string) *R2 {
	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String("https://" + accountID + ".r2.cloudflarestorage.com"),
		Credentials:  credentials.NewStaticCredentialsProvider(keyID, secret, ""),
		UsePathStyle: true,
	})
	return &R2{client: client, bucket: bucket, baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *R2) Put(ctx context.Context, key, contentType string, data []byte) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		// 名前に UUID が入っていて中身は変わらないので、長く覚えてもらう
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return "", err
	}
	return s.baseURL + "/" + key, nil
}

func (s *R2) Delete(ctx context.Context, url string) error {
	key, ok := keyOf(s.baseURL, url)
	if !ok {
		return nil
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	return err
}

func (s *R2) Owns(url string) bool {
	_, ok := keyOf(s.baseURL, url)
	return ok
}
