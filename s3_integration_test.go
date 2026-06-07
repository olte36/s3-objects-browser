//go:build integration

package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// awsIntegrationPreviewLimit bounds object preview reads during AWS integration tests.
const awsIntegrationPreviewLimit int64 = 1024

// awsIntegrationTimeout bounds live AWS calls during integration tests.
const awsIntegrationTimeout = 30 * time.Second

// TestAWSIntegration verifies the MinIO-backed service can browse a real AWS S3 bucket.
func TestAWSIntegration(t *testing.T) {
	if os.Getenv("S3BROWSER_AWS_INTEGRATION") != "1" {
		t.Skip("set S3BROWSER_AWS_INTEGRATION=1 to run AWS S3 integration test")
	}

	bucket := os.Getenv("S3BROWSER_AWS_BUCKET")
	key := os.Getenv("S3BROWSER_AWS_KEY")
	if bucket == "" || key == "" {
		t.Fatal("S3BROWSER_AWS_BUCKET and S3BROWSER_AWS_KEY must be set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), awsIntegrationTimeout)
	defer cancel()

	endpoint, err := parseEndpoint("aws")
	if err != nil {
		t.Fatalf("parse AWS endpoint: %v", err)
	}

	auth, err := newCredentialConfig(ctx, "aws", "", "", "")
	if err != nil {
		t.Fatalf("create AWS credential config: %v", err)
	}

	service, err := newMinioService(endpoint, auth)
	if err != nil {
		t.Fatalf("create MinIO service: %v", err)
	}

	buckets, err := service.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("list buckets: %v", err)
	}
	if !containsBucket(buckets, bucket) {
		t.Fatalf("bucket %q not found in ListBuckets response", bucket)
	}

	prefix := objectParentPrefix(key)
	progressCalls := 0
	objects, err := service.ListObjects(ctx, bucket, prefix, func(int) {
		progressCalls++
	})
	if err != nil {
		t.Fatalf("list objects for bucket %q prefix %q: %v", bucket, prefix, err)
	}
	if !containsObject(objects, key) {
		t.Fatalf("object %q not found in ListObjects response for prefix %q", key, prefix)
	}
	if progressCalls == 0 {
		t.Fatal("progress callback was not called while listing objects")
	}

	detail, err := service.InspectObject(ctx, bucket, key, awsIntegrationPreviewLimit)
	if err != nil {
		t.Fatalf("inspect object %q: %v", key, err)
	}
	if detail.Object.Key != key {
		t.Fatalf("inspected key = %q, want %q", detail.Object.Key, key)
	}
	if detail.Object.Size < 0 {
		t.Fatalf("inspected object size = %d, want non-negative", detail.Object.Size)
	}
	if detail.PreviewLen > awsIntegrationPreviewLimit {
		t.Fatalf("preview length = %d, want at most %d", detail.PreviewLen, awsIntegrationPreviewLimit)
	}
	if detail.Metadata == nil {
		t.Fatal("metadata map is nil")
	}
}

// containsBucket reports whether buckets includes name.
func containsBucket(buckets []bucketItem, name string) bool {
	for _, bucket := range buckets {
		if bucket.Name == name {
			return true
		}
	}
	return false
}

// containsObject reports whether objects includes key.
func containsObject(objects []objectItem, key string) bool {
	for _, object := range objects {
		if object.Key == key {
			return true
		}
	}
	return false
}

// objectParentPrefix returns the S3 prefix needed to list key's immediate parent.
func objectParentPrefix(key string) string {
	key = strings.TrimPrefix(key, "/")
	index := strings.LastIndex(key, "/")
	if index < 0 {
		return ""
	}
	return key[:index+1]
}
