package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

const propertiesFilename = "tags.properties"
const jsonFilename = "tags.json"
const fileMode = 0644

func main() {
	outDirParam := flag.String("out-dir", "/etc/config/", "output directory")
	instanceIDParam := flag.String("instance-id", "", "aws instance id")

	flag.Parse()

	instanceID, err := getInstanceID(*instanceIDParam)
	check(err, "Error getting instance")

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("eu-west-1"),
		config.WithSharedConfigProfile("deployTools"),
	)
	check(err, "Error loading config")

	asgClient := autoscaling.NewFromConfig(cfg)

	asgTags, err := getTagsFromASG(asgClient, instanceID)

	var tags map[string]string
	if (err != nil) {
		// check if err is lack of permission??
		//check(err, "Error getting tags from ASG")
		logf("Failed to get tags from ASG, falling back to EC2")
		ec2Client := ec2.NewFromConfig(cfg)

		tags, err = getTagsFromInstance(ec2Client, instanceID)

		check(err, "Error getting tags from EC2")
	} else {
		logf("Tags fetched from ASG")
		tags = asgTags
	}

	var fileContent string

	for key, value := range tags {
		fileContent = fileContent + fmt.Sprintf("%s=%s\n", key, value)
	}

	tagJSON, err := json.Marshal(tags)
	check(err, "not json")

	err = os.MkdirAll(*outDirParam, os.ModePerm)
	check(err, "couldn't create directory")

	propertiesPath := path.Join(*outDirParam, propertiesFilename)
	jsonPath := path.Join(*outDirParam, jsonFilename)

	err = ioutil.WriteFile(propertiesPath, []byte(fileContent), fileMode)
	check(err, "couldn't create properties file")

	err = ioutil.WriteFile(jsonPath, tagJSON, fileMode)
	check(err, "couldn't create JSON file")

	logf("Written %d tags to %s and %s", len(tags), propertiesPath, jsonPath)
}

func getInstanceID(instanceIDParam string) (string, error) {
	if instanceIDParam != "" {
		logf("Instance ID %s passed as param", instanceIDParam)
		return instanceIDParam, nil
	}

	client := imds.New(imds.Options{})
	input := imds.GetInstanceIdentityDocumentInput{}
	output, err := client.GetInstanceIdentityDocument(context.TODO(), &input)

	if err != nil {
		return "", err
	}

	logf("Instance ID from IMDS is %s", output.InstanceID)

	return output.InstanceID, nil
}

func getTagsFromASG(client *autoscaling.Client, instanceID string) (map[string]string, error) {
	response := map[string]string{}

	describeASGInstancesInput := autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []string{instanceID},
	}
	describeASGInstancesOutput, err := client.DescribeAutoScalingInstances(context.TODO(), &describeASGInstancesInput)

	if err != nil {
		return response, err
	}

	if asgInstancesLength := len(describeASGInstancesOutput.AutoScalingInstances); asgInstancesLength != 1 {
		return response, fmt.Errorf("Expected 1 AutoScalingInstances. Got %v", asgInstancesLength)
	}

	asgName := describeASGInstancesOutput.AutoScalingInstances[0].AutoScalingGroupName

	describeASGInput := autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{*asgName},
	}
	describeASGOutput, err := client.DescribeAutoScalingGroups(context.TODO(), &describeASGInput)

	if err != nil {
		return response, err
	}

	if asgLength := len(describeASGOutput.AutoScalingGroups); asgLength != 1 {
		return response, fmt.Errorf("Expected 1 AutoScalingGroups. Got %v", asgLength)
	}

	for _, tag := range describeASGOutput.AutoScalingGroups[0].Tags {
		response[*tag.Key] = *tag.Value
	}

	return response, nil
}

func getTagsFromInstance(client *ec2.Client, instanceID string) (map[string]string, error) {
	response := map[string]string{}

	describeEC2InstanceInput := ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	describeEC2InstanceOutput, err := client.DescribeInstances(context.TODO(), &describeEC2InstanceInput)
	
	if err != nil {
		return response, err
	}

	if reservationsLength := len(describeEC2InstanceOutput.Reservations); reservationsLength != 1 {
		return response, fmt.Errorf("Expected 1 Reservation. Got %v", reservationsLength)
	}

	instances := describeEC2InstanceOutput.Reservations[0].Instances

	if length := len(instances); length != 1 {
		return response, fmt.Errorf("Expected 1 Instance. Got %v", length)
	}

	for _, tag := range instances[0].Tags {
		response[*tag.Key] = *tag.Value
	}

	return response, nil
}

func logf(msg string, v ...interface{}) {
	prefixedMessage := fmt.Sprintf("[instance-tag-discovery] %s", msg)
	log.Printf(prefixedMessage, v...)
}

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("[instance-tag-discovery] %s; %v", msg, err)
	}
}
