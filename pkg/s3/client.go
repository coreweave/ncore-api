package s3

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewClient(host string) *S3Svc {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithDefaultRegion("default"),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: host}, nil
				})))
	if err != nil {
		fmt.Println("Couldn't load default configuration. Ensure Env Var's are set for access and secret key.")
		fmt.Println(err)
		return nil
	}

	s3Client := s3.NewFromConfig(sdkConfig)
	return &S3Svc{Client: s3Client}
}

type S3Svc struct {
	Client *s3.Client
}

func (svc *S3Svc) GetObject(
	bucketName string, objectKey string) ([]byte, error) {
	request, err := svc.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.Options) {
	})
	if err != nil {
		log.Printf("Couldn't get GetObjectOutput for %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
		return nil, err
	}
	defer request.Body.Close()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
		return nil, err
	}
	return body, err
}

func NewPresigner(s S3Svc) HttpPresigner {

	preSignClient := s3.NewPresignClient(s.Client)

	return &Presigner{preSignClient}
}

//go:generate mockgen --build_flags=--mod=mod -package s3 -destination mock_s3_test.go . HttpPresigner
type HttpPresigner interface {
	GetObject(bucketName string, objectKey string, lifetimeSecs int64) (*v4.PresignedHTTPRequest, error)
}

// Presigner To request a presigned object from simply request GetObject
//
//	request, err := myPresigner.GetObject(bucket,key,3600)
//	httpRequest := http.NewRequest(request.URL, "GET", nil)
type Presigner struct {
	PresignClient *s3.PresignClient
}

// GetObject makes a presigned request that can be used to get an object from a bucket.
// The presigned request is valid for the specified number of seconds.
func (presigner Presigner) GetObject(
	bucketName string, objectKey string, lifetimeSecs int64) (*v4.PresignedHTTPRequest, error) {
	request, err := presigner.PresignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		log.Printf("Couldn't get a presigned request to get %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return request, err
}
