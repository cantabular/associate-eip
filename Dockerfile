FROM golang:1

RUN go get -v github.com/awslabs/aws-sdk-go/aws

COPY . /go/src/github.com/pwaller/associate-eip

RUN go install github.com/pwaller/associate-eip