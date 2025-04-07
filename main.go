package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func Metadata(imdsClient *imds.Client, path string) (string, error) {
	metadataOutput, err := imdsClient.GetMetadata(context.Background(), &imds.GetMetadataInput{Path: path})
	if err != nil {
		return "", err
	}
	defer metadataOutput.Content.Close()

	metadataOutputBytes, err := io.ReadAll(metadataOutput.Content)
	if err != nil {
		return "", err
	}
	return string(metadataOutputBytes), nil
}

// Returns current IP address via metadata endpoint or "" on error.
func MyIP(imdsClient *imds.Client) string {
	ip, _ := Metadata(imdsClient, "public-ipv4")
	return ip
}

func WaitForIP(imdsClient *imds.Client, target string) bool {

	const MAX = 120 // Maximum number of seconds to wait

	for i := 0; i < MAX; i++ {
		ip := MyIP(imdsClient)
		if ip == target {
			log.Printf("IP updated!: %q", ip)
			return true
		}
		log.Printf("Waiting for IP address update: %q", ip)
		time.Sleep(1 * time.Second)
	}
	return false
}

func ThisInstanceID(imdsClient *imds.Client) (string, error) {
	return Metadata(imdsClient, "instance-id")
}

func ThisAvailabilityZone(imdsClient *imds.Client) (string, error) {
	result, err := Metadata(imdsClient, "placement/availability-zone")
	if err == nil {
		result = result[:len(result)-1]
	}
	return result, err
}

func main() {
	log.SetFlags(0)

	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatalf("usage: %s <PublicIP>", os.Args[0])
	}
	publicIP := args[0]

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Unable to load AWS config: %v", err)
	}

	imdsClient := imds.NewFromConfig(cfg)

	thisInstanceID, err := ThisInstanceID(imdsClient)
	if err != nil {
		log.Fatalf("Unable to determine instance id: %v", err)
	}

	thisAvailabilityZone, err := ThisAvailabilityZone(imdsClient)
	if err != nil {
		log.Fatalf("Unable to determine availability zone: %v", err)
	}

	log.Println("InstanceID:", thisInstanceID, "AZ:", thisAvailabilityZone)

	cfg.Region = thisAvailabilityZone

	svc := ec2.NewFromConfig(cfg)

	desc, err := svc.DescribeAddresses(context.Background(), &ec2.DescribeAddressesInput{
		PublicIps: []string{(publicIP)},
	})

	if err != nil {
		log.Fatalf("Unable to describe EIPs: %v", err)
	}

	if len(desc.Addresses) != 1 {
		log.Fatalf("Expected exactly 1 address, got %v", len(desc.Addresses))
	}

	allocation := desc.Addresses[0]

	resp, err := svc.AssociateAddress(context.Background(), &ec2.AssociateAddressInput{
		InstanceId:         &thisInstanceID,
		AllowReassociation: aws.Bool(true),
		AllocationId:       allocation.AllocationId,
	})

	if err != nil {
		log.Fatalf("Unable to associate allocation: %v", err)
	}

	log.Println("Associated:", *resp.AssociationId)

	if !WaitForIP(imdsClient, publicIP) {
		log.Fatal("Failed to see public IP update in a timely fashion.")
	}
}
