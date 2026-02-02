package registry

import (
	"context"
	"fmt"
	"io"
)

type s3Registry struct {
	bucket   string
	region   string
	endpoint string
}

func NewS3Registry(bucket, region, endpoint string) (Registry, error) {
	// TODO: Implement S3 registry with AWS SDK v2
	// For now, return a stub that indicates it's not yet implemented
	return &s3Registry{
		bucket:   bucket,
		region:   region,
		endpoint: endpoint,
	}, nil
}

func (r *s3Registry) Push(ctx context.Context, name, tag string, data io.Reader) error {
	return fmt.Errorf("S3 registry not yet fully implemented - use filesystem registry for now")
}

func (r *s3Registry) Pull(ctx context.Context, name, tag string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("S3 registry not yet fully implemented - use filesystem registry for now")
}

func (r *s3Registry) Exists(ctx context.Context, name, tag string) (bool, error) {
	return false, fmt.Errorf("S3 registry not yet fully implemented - use filesystem registry for now")
}
