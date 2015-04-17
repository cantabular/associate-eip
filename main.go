package main

import (
	"fmt"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/aws/awsutil"
	"github.com/awslabs/aws-sdk-go/service/ec2"
)

func main() {

	svc := ec2.New(nil)

	desc, err := svc.DescribeAddresses(&ec2.DescribeAddressesInput{
		PublicIPs: []*string{aws.String("52.16.160.41")},
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(awsutil.StringValue(desc.Addresses))
	// svc.AssociateAddress(&ec2.AssociateAddressInput{})
}
