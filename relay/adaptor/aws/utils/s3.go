package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
)

// UploadBase64ImageToS3 uploads a base64 encoded image to S3 and returns the URL
func UploadBase64ImageToS3(ctx context.Context, s3Client *s3.Client, bucket, endpoint, base64Data string) (string, error) {
	// Remove data URL prefix if present
	if len(base64Data) > 0 && base64Data[0] == 'd' {
		// Check if it starts with "data:image/"
		if len(base64Data) > 11 && base64Data[:11] == "data:image/" {
			// Find the comma that separates the metadata from the base64 data
			for i := 11; i < len(base64Data); i++ {
				if base64Data[i] == ',' {
					base64Data = base64Data[i+1:]
					break
				}
			}
		}
	}

	// Decode base64 data
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}
	// get mime type from base64 data
	mimeType := mimetype.Detect(imageData)
	// Generate a unique filename
	filename := fmt.Sprintf("images/%d-%s%s", time.Now().UnixNano(), random.GetUUID(), mimeType.Extension())

	// Upload to S3
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(mimeType.String()),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate the URL
	url := fmt.Sprintf("https://%s.%s/%s", bucket, endpoint, filename)
	logger.Infof(ctx, "Successfully uploaded image to S3: %s", url)
	return url, nil
}

func UploadImageFromUrlToS3(ctx context.Context, s3Client *s3.Client, bucket, endpoint, url string) (string, error) {
	// Download the image from the URL
	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download image from URL: %w", err)
	}
	defer response.Body.Close()

	// get mime type from response
	statusCode := response.StatusCode
	if statusCode != 200 {
		return "", fmt.Errorf("failed to get image type from URL")
	}

	// get image data from response
	imageData, err := io.ReadAll(response.Body)
	mimeType := mimetype.Detect(imageData)
	filename := fmt.Sprintf("images/%d-%s%s", time.Now().UnixNano(), random.GetUUID(), mimeType.Extension())
	if err != nil {
		return "", fmt.Errorf("failed to read image data from URL: %w", err)
	}

	// upload image data to S3
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(mimeType.String()),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image to S3: %w", err)
	}

	// generate the url
	imageUrl := fmt.Sprintf("https://%s.%s/%s", bucket, endpoint, filename)
	logger.Infof(ctx, "Successfully uploaded image to S3: %s", url)
	return imageUrl, nil
}
