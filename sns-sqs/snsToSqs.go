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
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

func postHandler(w http.ResponseWriter, r *http.Request) {
	sess := session.Must(session.NewSession())
	log.Printf("This is request %+v\n", r)
	by, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	snsClient := sns.New(sess)
	topicArn := "arn:aws:sns:ap-south-1:306488905853:new-test-topic"

	_, err = snsClient.Publish(&sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(by)),
	})
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	//ctx := context.Background()
	//lambda.StartWithOptions(sqsToS3, lambda.WithContext(ctx))
	//if ctx.Value("error") != nil || ctx.Err() != nil {
	//	log.Println(err)
	//	json.NewEncoder(w).Encode(err.Error())
	//	return
	//}
	//log.Printf("this is ctx value vatsal %s", ctx.Value("id"))
	//id, err := json.Marshal(map[string]interface{}{"id": ctx.Value("id")})
	//if err != nil {
	//	log.Println(err)
	//	json.NewEncoder(w).Encode(err.Error())
	//	return
	//}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return
}
func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	s3Client := s3.New(session.Must(session.NewSession()))
	resp, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("test-bucket-write"),
		Key:    aws.String(fmt.Sprintf("%s.txt", id)),
	})
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
	//json.NewEncoder(w).Encode(string(body))
}

func router() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/sns", postHandler).Methods(http.MethodPost)
	r.HandleFunc("/sns/{id}", getHandler).Methods(http.MethodGet)
	return r
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r := router()
	reqPath := request.Path
	if request.Path == "" {
		reqPath = "/sns"
	}
	httpReq, err := http.NewRequest(request.HTTPMethod, reqPath, bytes.NewReader([]byte(request.Body)))
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}
	w := core.NewProxyResponseWriter()
	r.ServeHTTP(w, httpReq)

	// Create an APIGatewayProxyResponse from the http.Response
	resp, err := w.GetProxyResponse()
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
