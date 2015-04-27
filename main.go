package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/ec2"
)

func Metadata(path string) (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/" + path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Returns current IP address via metadata endpoint or "" on error.
func MyIP() string {
	ip, _ := Metadata("public-ipv4")
	return ip
}

func WaitForIP(target string) bool {

	const MAX = 120 // Maximum number of seconds to wait

	for i := 0; i < MAX; i++ {
		ip := MyIP()
		if ip == target {
			log.Printf("IP updated!: %q", ip)
			return true
		}
		log.Printf("Waiting for IP address update: %q", ip)
		time.Sleep(1 * time.Second)
	}
	return false
}

func ThisInstanceID() (string, error) {
	return Metadata("instance-id")
}

func ThisAvailabilityZone() (string, error) {
	result, err := Metadata("placement/availability-zone")
	if err == nil {
		result = result[:len(result)-1]
	}
	return result, err
}

func main() {
	log.SetFlags(0)

	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatal("usage: %s <PublicIP>", os.Args[0])
	}
	publicIP := args[0]

	thisInstanceID, err := ThisInstanceID()
	if err != nil {
		log.Fatalf("Unable to determine instance id: %v", err)
	}

	thisAvailabilityZone, err := ThisAvailabilityZone()
	if err != nil {
		log.Fatalf("Unable to determine availability zone: %v", err)
	}

	log.Println("InstanceID:", thisInstanceID, "AZ:", thisAvailabilityZone)

	svc := ec2.New(&aws.Config{
		Region: thisAvailabilityZone,
	})

	desc, err := svc.DescribeAddresses(&ec2.DescribeAddressesInput{
		PublicIPs: []*string{aws.String(publicIP)},
	})

	if err != nil {
		log.Fatalf("Unable to describe EIPs: %v", err)
	}

	if len(desc.Addresses) != 1 {
		log.Fatalf("Expected exactly 1 address, got %v", len(desc.Addresses))
	}

	allocation := desc.Addresses[0]

	resp, err := svc.AssociateAddress(&ec2.AssociateAddressInput{
		InstanceID:         &thisInstanceID,
		AllowReassociation: aws.Boolean(true),
		AllocationID:       allocation.AllocationID,
	})

	if err != nil {
		log.Fatalf("Unable to associate allocation: %v", err)
	}

	log.Println("Associated:", *resp.AssociationID)

	if !WaitForIP(publicIP) {
		log.Fatal("Failed to see public IP update in a timely fashion.")
	}
}
