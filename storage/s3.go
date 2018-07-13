package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
)

type S3Client struct {
	client  *s3.S3
	session *session.Session
}

func NewS3Client(regionName string) *S3Client {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(regionName),
	}))

	svc := s3.New(sess)

	return &S3Client{svc, sess}
}

func (s *S3Client) HeadObject(bucketName, key string) (map[string]*string, error) {
	headObjectInput := s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	res, err := s.client.HeadObject(&headObjectInput)
	return res.Metadata, err
}

func (s *S3Client) PutObject(bucketName, key string, body io.ReadSeeker) error {
	putObjectInput := s3.PutObjectInput{
		Body:   body,
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	_, err := s.client.PutObject(&putObjectInput)

	return err
}

func (s *S3Client) GetObject(bucketName, key string) (io.ReadCloser, error) {
	getObjectInput := s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	output, err := s.client.GetObject(&getObjectInput)
	return output.Body, err
}