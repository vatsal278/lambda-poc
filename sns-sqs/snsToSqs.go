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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

type Data struct {
	Id   string `json:"id"`
	Data string `json:"data"`
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	sess := session.Must(session.NewSession())

	by, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	var data = Data{Id: uuid.NewString(), Data: string(by)}
	byt, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	snsClient := sns.New(sess)
	topicArn := "arn:aws:sns:ap-south-1:306488905853:create_file_s3"

	_, err = snsClient.Publish(&sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(byt)),
	})
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(data.Id))
	return
}
func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	s3Client := s3.New(session.Must(session.NewSession()))
	resp, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("s3-data-bucket0"),
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
	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data.Data))
}

func router() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", postHandler).Methods(http.MethodPost)
	r.HandleFunc("/{id}", getHandler).Methods(http.MethodGet)
	return r
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r := router()
	reqPath := request.Path
	if request.Path == "" {
		reqPath = "/"
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
