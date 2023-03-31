package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"log"
)

type Data struct {
	id   string
	data []byte
}

func sqsToS3() error {
	sess := session.Must(session.NewSession())
	sqsClient := sqs.New(sess)

	queueUrl := "https://sqs.ap-south-1.amazonaws.com/306488905853/my-sqs"
	result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds:     aws.Int64(20),
	})
	if err != nil {
		log.Println(err)
		return err
	}
	if len(result.Messages) == 0 {
		log.Println("no messages in the queue")
		return nil
	}

	message := *result.Messages[0].Body
	var data Data
	err = json.Unmarshal([]byte(message), &data)
	if err != nil {
		log.Println(err)
		return err
	}

	s3Client := s3.New(session.Must(session.NewSession()))

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("test-bucket-write"),
		Key:    aws.String(fmt.Sprintf("%s.txt", data.id)),
		Body:   bytes.NewReader(data.data),
	})
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: result.Messages[0].ReceiptHandle,
	})
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to delete message from SQS: %v", err)
	}
	return nil
}

func main() {
	lambda.Start(sqsToS3)
}
