package wasabi

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Wasabi configuration
var (
	wasabiEndpoint = "https://s3.ap-southeast-1.wasabisys.com" // Adjust to your Wasabi region
	region         = "ap-southeast-1"                          // Wasabi region
	bucketName     = "jayden-bucket-testing"
	accessKey      = "4B4JQOKZJVVAOHDX9NLV"
	secretKey      = "8eJTXOGqzu3oD3iJFDkMb4FeRKjdwEaOBNm9XlUe"
)

var (
	awsConfig            *aws.Config
	parsedWasabiEndpoint *url.URL
)

func IsValid() error {
	if awsConfig == nil {
		return fmt.Errorf("invalid sdk config")
	}

	if parsedWasabiEndpoint == nil {
		return fmt.Errorf("invalid Wasabi endpoint url")
	}

	return nil
}

func LoadConfig() error {
	// Load config normally (region and credentials)
	cfg, ConfigErr := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)

	if ConfigErr != nil {
		return fmt.Errorf("unable to load SDK config, %v", ConfigErr)
	}

	// Parse Wasabi endpoint URL
	parsedEndpoint, ParseErr := url.Parse(wasabiEndpoint)
	if ParseErr != nil {
		return fmt.Errorf("invalid Wasabi endpoint url: %v", ParseErr)
	}

	awsConfig = &cfg
	parsedWasabiEndpoint = parsedEndpoint

	return nil
}

func UploadFile(file multipart.File, handler *multipart.FileHeader, UUID string) error {

	if err := IsValid(); err != nil {
		return err
	}

	if file == nil {
		return fmt.Errorf("failed to upload file to wasabi, file is nil")
	}

	if handler == nil {
		return fmt.Errorf("failed to upload file to wasabi, handler is nil")
	}

	// Create S3 client with service-specific endpoint override
	client := s3.NewFromConfig(*awsConfig, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL(parsedWasabiEndpoint.String())
		// Mark the endpoint as immutable (won't be changed by SDK)
		o.UsePathStyle = true // Wasabi recommends path-style addressing
	})

	fileSize := handler.Size           // int64
	fileSizePtr := aws.Int64(fileSize) // convert to *int64 using aws helper

	_, UploadErr := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        &bucketName,
		Key:           aws.String(UUID),
		Body:          file,
		ContentLength: fileSizePtr,
		ContentType:   aws.String("image/jpeg"),
		ACL:           "public-read",
	})

	if UploadErr != nil {
		return fmt.Errorf("failed to upload image: %v", UploadErr)
	}

	return nil
}

func DeleteFile(key string) error {

	if err := IsValid(); err != nil {
		return err
	}

	// Create S3 client
	client := s3.NewFromConfig(*awsConfig, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL(parsedWasabiEndpoint.String())
		o.UsePathStyle = true
	})

	// Delete object
	_, DeleteErr := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if DeleteErr != nil {
		return fmt.Errorf("unable to delete object: %v", DeleteErr)
	}

	return nil
}

func DownloadImage(Key, filepath string) error {

	if err := IsValid(); err != nil {
		return err
	}

	// Custom endpoint
	client := s3.NewFromConfig(*awsConfig, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL(parsedWasabiEndpoint.String())
		o.UsePathStyle = true
	})

	// Get the object
	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(Key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object: %v", err)
	}
	defer resp.Body.Close()

	// Write to file
	outFile, err := os.Create(filepath + Key)

	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	return nil
}
