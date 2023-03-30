package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"io/ioutil"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func readFromS3(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Fetch the query and path parameters from the request
	queryParams := event.QueryStringParameters
	pathParams := event.PathParameters

	// Print the query and path parameters
	fmt.Println("Query parameters:")
	for key, value := range queryParams {
		fmt.Printf("%s: %s\n", key, value)
	}

	fmt.Println("Path parameters:")
	for key, value := range pathParams {
		fmt.Printf("%s: %s\n", key, value)
	}
	s3Client := s3.New(session.Must(session.NewSession()))
	resp, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("test-bucket-write"),
		Key:    aws.String("my-file.txt"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}
	log.Print(string(body))
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(readFromS3)
}

//aws lambda invoke --function-name writeToS3 --payload '{"key1":"value1", "key2":"value2", "key4":"value3"}' response.json
