https://www.youtube.com/watch?v=Czny2I2uGJA

https://www.golinuxcloud.com/golang-aws-lambda/

https://aws.amazon.com/developer/language/go/

https://github.com/aws/aws-lambda-go

http://www.inanzzz.com/index.php/post/yrut/handling-multiple-aws-lambda-handlers-in-golang-using-localstack

https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html

https://repost.aws/knowledge-center/lambda-execution-role-s3-bucket
build using  --> 
```
GOOS=linux GOARCH=amd64 go build -o write/main write/writeToS3.go
```

create func using -->
```
aws lambda create-function     --function-name writeS3     --runtime go1.x     --handler main   --zip-file fileb://./write/main.zip.zip --publish --role arn:aws:iam::306488905853:role/service-role/test-lambda-role2
```
