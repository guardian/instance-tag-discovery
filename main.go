package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
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

	client := autoscaling.NewFromConfig(cfg)

	tags, err := getTagsFromASG(client, instanceID)

	check(err, "Error getting tags")

	var fileContent string

	for key, value := range tags {
		fileContent = fileContent + fmt.Sprintf("%s=%s\n", key, value)
	}

	tagJSON, err := json.Marshal(tags)
	check(err, "not json")

	ioutil.WriteFile(path.Join(*outDirParam, propertiesFilename), []byte(fileContent), fileMode)
	ioutil.WriteFile(path.Join(*outDirParam, jsonFilename), tagJSON, fileMode)
}

func getInstanceID(instanceIDParam string) (string, error) {
	if instanceIDParam != "" {
		return instanceIDParam, nil
	}

	client := imds.New(imds.Options{})
	input := imds.GetInstanceIdentityDocumentInput{}
	output, err := client.GetInstanceIdentityDocument(context.TODO(), &input)

	if err != nil {
		return "", err
	}

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

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s; %v", msg, err)
	}
}
