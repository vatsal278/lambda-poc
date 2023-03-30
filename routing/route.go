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
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

func postHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("This is request %+v\n", r)
	by, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	key := uuid.NewString()
	s3Client := s3.New(session.Must(session.NewSession()))
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("test-bucket-write"),
		Key:    aws.String(fmt.Sprintf("%s.txt", key)),
		Body:   bytes.NewReader(by),
	})
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	id, err := json.Marshal(map[string]string{"id": key})
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(id)
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
