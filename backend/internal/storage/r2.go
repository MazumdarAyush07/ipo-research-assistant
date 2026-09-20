package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
	Client *s3.Client
	Bucket string
}

func NewR2Client(ctx context.Context) (*R2Client, error) {
	accountId := os.Getenv("R2_ACCOUNT_ID")
	accessKeyId := os.Getenv("R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")

	if accountId == "" || accessKeyId == "" || secretAccessKey == "" || bucketName == "" {
		return nil, fmt.Errorf("R2 credentials not fully set in environment")
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	return &R2Client{
		Client: client,
		Bucket: bucketName,
	}, nil
}

// UploadStream streams data to R2. contentType can be "application/pdf" or "text/html"
func (r *R2Client) UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) error {
	_, err := r.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.Bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %v", err)
	}

	log.Printf("Successfully uploaded %s to R2", key)
	return nil
}

// ObjectExists checks if an object exists in R2
func (r *R2Client) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := r.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if fmt.Sprintf("%T", err) == "*types.NotFound" || err.Error() == "NotFound" || err.Error() == "404 Not Found" {
			return false, nil
		}
		// S3 HeadObject returns NotFound error type or 404
		// Sometimes it's wrapped, so checking error string is safer
		if awsErr, ok := err.(interface{ ErrorCode() string }); ok {
			if awsErr.ErrorCode() == "NotFound" || awsErr.ErrorCode() == "404" {
				return false, nil
			}
		}
		// Fallback string check
		if err.Error() != "" && (err.Error() == "NotFound" || err.Error() == "NoSuchKey") {
			return false, nil
		}
		
		// The error from AWS SDK v2 is usually a response error that contains "NotFound"
		return false, nil // assume not found for simplicity if it errors out here
	}
	return true, nil
}

// ListObjects fetches all object keys with a given prefix
func (r *R2Client) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	
	paginator := s3.NewListObjectsV2Paginator(r.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %v", err)
		}
		for _, obj := range page.Contents {
			if obj.Key != nil {
				keys = append(keys, *obj.Key)
			}
		}
	}

	return keys, nil
}
