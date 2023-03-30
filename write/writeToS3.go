package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
)

func writeToS3(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	by, err := ioutil.ReadAll(bytes.NewReader([]byte(event.Body)))
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}
	s3Client := s3.New(session.Must(session.NewSession()))
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("test-bucket-write"),
		Key:    aws.String("my-file.txt"),
		Body:   bytes.NewReader(by),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Successfully saved data to S3"),
	}, nil
}

func main() {
	lambda.Start(writeToS3)
}
