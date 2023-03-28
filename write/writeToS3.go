package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"log"
)

type RequestData struct {
	Message string `json:"message"`
}

func writeToS3(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//lc, _ := lambdacontext.FromContext(ctx)
	//log.Printf("this is the context for aws lambda %+v \n", lc)
	//log.Printf("this is the event details %+v \n", event)
	//body := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(event.Body)))
	by, err := ioutil.ReadAll(bytes.NewReader([]byte(event.Body)))
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}
	log.Println(string(by))
	var requestData RequestData
	if err := json.Unmarshal(by, &requestData); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}
	log.Printf("this is the request body%+v \n", requestData)
	s3Client := s3.New(session.Must(session.NewSession()))
	buffer := bytes.NewReader([]byte(requestData.Message))
	_, err = s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String("test-bucket-write"),
		Key:    aws.String("my-file.txt"),
		Body:   buffer,
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
