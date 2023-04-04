package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"log"
	"os"
)

type Data struct {
	Id   string `json:"id"`
	Data string `json:"data"`
}

func sqsToS3() error {
	sess := session.Must(session.NewSession())
	sqsClient := sqs.New(sess)
	s3Client := s3.New(sess)
	queueUrl := os.Getenv("sqsUrl")
	result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int64(10),
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
	for _, msg := range result.Messages {
		message := *msg.Body
		var snsData events.SNSEntity
		err = json.Unmarshal([]byte(message), &snsData)
		if err != nil {
			log.Println(err)
			return err
		}

		var data Data
		err = json.Unmarshal([]byte(snsData.Message), &data)
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String("s3-data-bucket0"),
			Key:    aws.String(fmt.Sprintf("%s.txt", data.Id)),
			Body:   bytes.NewReader([]byte(data.Data)),
		})
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueUrl),
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			log.Println(err)
			return fmt.Errorf("failed to delete message from SQS: %v", err)
		}
	}
	return nil
}

func main() {
	lambda.Start(sqsToS3)
}
