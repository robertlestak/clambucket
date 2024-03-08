package object

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func ParseS3Uri(s3Uri string) (string, string, error) {
	// ensure the S3 URI has the s3:// prefix
	if len(s3Uri) < 5 || s3Uri[0:5] != "s3://" {
		return "", "", fmt.Errorf("invalid S3 URI: %s", s3Uri)
	}
	// remove the s3:// prefix
	s3Uri = s3Uri[5:]

	// split the S3 URI into its parts
	parts := strings.SplitN(s3Uri, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid S3 URI: %s", s3Uri)
	}

	return parts[0], parts[1], nil
}

func S3Get(s3Uri, localDir, roleArn string) error {
	// parse the S3 URI into its parts
	bucketName, objectKey, err := ParseS3Uri(s3Uri)
	if err != nil {
		return err
	}
	// Create a new AWS session
	sess := session.Must(session.NewSession())
	var svc *s3.S3
	if roleArn != "" {
		// Assume the specified IAM role
		creds := stscreds.NewCredentials(sess, roleArn)
		// Create a new S3 service client with the assumed role credentials
		svc = s3.New(sess, &aws.Config{Credentials: creds})
	} else {
		// Create a new S3 service client
		svc = s3.New(sess)
	}
	// Create a new GetObjectInput
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	// Get the object from S3
	resp, err := svc.GetObject(input)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fp := filepath.Join(localDir, objectKey)
	// Create a new file to write the object contents to
	file, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the object contents to the file
	_, err = file.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func S3Put(localFile, s3Uri, roleArn string) error {
	// parse the S3 URI into its parts
	bucketName, objectKey, err := ParseS3Uri(s3Uri)
	if err != nil {
		return err
	}
	// Create a new AWS session
	sess := session.Must(session.NewSession())
	var svc *s3.S3
	if roleArn != "" {
		// Assume the specified IAM role
		creds := stscreds.NewCredentials(sess, roleArn)
		// Create a new S3 service client with the assumed role credentials
		svc = s3.New(sess, &aws.Config{Credentials: creds})
	} else {
		// Create a new S3 service client
		svc = s3.New(sess)
	}
	// Create a new PutObjectInput
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	// Open the local file
	file, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set the object contents
	input.Body = file

	// Put the object to S3
	_, err = svc.PutObject(input)
	if err != nil {
		return err
	}

	return nil
}

func S3Copy(srcS3Uri, dstS3Uri, roleArn string) error {
	// parse the source S3 URI into its parts
	srcBucketName, srcObjectKey, err := ParseS3Uri(srcS3Uri)
	if err != nil {
		return err
	}
	// parse the destination S3 URI into its parts
	dstBucketName, dstObjectKey, err := ParseS3Uri(dstS3Uri)
	if err != nil {
		return err
	}
	// Create a new AWS session
	sess := session.Must(session.NewSession())
	var svc *s3.S3
	if roleArn != "" {
		// Assume the specified IAM role
		creds := stscreds.NewCredentials(sess, roleArn)
		// Create a new S3 service client with the assumed role credentials
		svc = s3.New(sess, &aws.Config{Credentials: creds})
	} else {
		// Create a new S3 service client
		svc = s3.New(sess)
	}
	// Create a new CopyObjectInput
	input := &s3.CopyObjectInput{
		CopySource: aws.String(srcBucketName + "/" + srcObjectKey),
		Bucket:     aws.String(dstBucketName),
		Key:        aws.String(dstObjectKey),
	}

	// Copy the object in S3
	_, err = svc.CopyObject(input)
	if err != nil {
		return err
	}

	return nil
}

func S3Delete(s3Uri, roleArn string) error {
	// parse the S3 URI into its parts
	bucketName, objectKey, err := ParseS3Uri(s3Uri)
	if err != nil {
		return err
	}
	// Create a new AWS session
	sess := session.Must(session.NewSession())
	var svc *s3.S3
	if roleArn != "" {
		// Assume the specified IAM role
		creds := stscreds.NewCredentials(sess, roleArn)
		// Create a new S3 service client with the assumed role credentials
		svc = s3.New(sess, &aws.Config{Credentials: creds})
	} else {
		// Create a new S3 service client
		svc = s3.New(sess)
	}
	// Create a new DeleteObjectInput
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	// Delete the object from S3
	_, err = svc.DeleteObject(input)
	if err != nil {
		return err
	}

	return nil
}
