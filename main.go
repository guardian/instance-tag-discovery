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
	"strings"

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
	instanceIDParam := flag.String("instance-id", "", "AWS instance id")
	profileParam := flag.String("profile", "deployTools", "AWS credentials profile (useful if running locally)")

	flag.Parse()

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("eu-west-1"),
		config.WithSharedConfigProfile(*profileParam),
	)
	check(err, "error loading config")

	_, err = cfg.Credentials.Retrieve(context.TODO())
	check(err, "unable to retrieve credentials")

	imdsClient := imds.New(imds.Options{})
	asgClient := autoscaling.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)

	instanceID, err := getInstanceID(imdsClient, *instanceIDParam)
	check(err, "error getting instance")

	tags, err := tagsFromIMDS(imdsClient)
	if err != nil {
		logf("failed to get tags from instance metadata, falling back to ASG. Metadata error was: %v", err)
		tags, err = getTagsFromASG(asgClient, instanceID)

		if err != nil {
			logf("failed to get tags from ASG, falling back to EC2. ASG error was: %v", err)
			tags, err = getTagsFromInstance(ec2Client, instanceID)
		}
	}

	tagProperties := ""
	for key, value := range tags {
		tagProperties = tagProperties + fmt.Sprintf("%s=%s\n", key, value)
	}

	tagJSON, err := json.Marshal(tags)
	check(err, "unable to marshall tags as JSON")

	err = os.MkdirAll(*outDirParam, os.ModePerm)
	check(err, "couldn't create out directory")

	propertiesPath := path.Join(*outDirParam, propertiesFilename)
	jsonPath := path.Join(*outDirParam, jsonFilename)

	err = ioutil.WriteFile(propertiesPath, []byte(tagProperties), fileMode)
	check(err, "couldn't create properties file")

	err = ioutil.WriteFile(jsonPath, tagJSON, fileMode)
	check(err, "couldn't create JSON file")

	logf("Written %d tags to %s and %s", len(tags), propertiesPath, jsonPath)
}

func getInstanceID(client *imds.Client, instanceIDParam string) (string, error) {
	if instanceIDParam != "" {
		logf("Instance ID %s passed as param", instanceIDParam)
		return instanceIDParam, nil
	}

	input := imds.GetInstanceIdentityDocumentInput{}
	output, err := client.GetInstanceIdentityDocument(context.TODO(), &input)
	if err != nil {
		return "", err
	}

	return output.InstanceID, nil
}

func tagsFromIMDS(client *imds.Client) (map[string]string, error) {
	tags := map[string]string{}

	input := imds.GetMetadataInput{Path: "tags/instance"}
	resp, err := client.GetMetadata(context.TODO(), &input)
	if err != nil {
		return tags, fmt.Errorf("unable to call tags metadata endpoint: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Content)
	defer resp.Content.Close()
	if err != nil {
		return tags, fmt.Errorf("unable to read body of imds tags request:  %w", err)
	}

	tagNames := strings.Split(strings.TrimSpace(string(body)), "\n")

	for _, name := range tagNames {
		input := imds.GetMetadataInput{Path: "tags/instance/" + name}
		value, _ := client.GetMetadata(context.TODO(), &input)
		if err == nil {
			body, _ = ioutil.ReadAll(value.Content)
			tags[name] = string(body)
			continue
		}
	}

	return tags, err
}

func getTagsFromASG(client *autoscaling.Client, instanceID string) (map[string]string, error) {
	response := map[string]string{}

	describeASGInstancesInput := autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []string{instanceID},
	}
	describeASGInstancesOutput, err := client.DescribeAutoScalingInstances(context.TODO(), &describeASGInstancesInput)

	if err != nil {
		return response, fmt.Errorf("unable to describe asg instances: %w", err)
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

	input := ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}}
	output, err := client.DescribeInstances(context.TODO(), &input)
	if err != nil {
		return response, err
	}

	reservationsLength := len(output.Reservations)
	if reservationsLength != 1 {
		return response, fmt.Errorf("expected 1 Reservation. Got %v", reservationsLength)
	}

	instances := output.Reservations[0].Instances
	if len(instances) != 1 {
		return response, fmt.Errorf("expected 1 Instance. Got %v", len(instances))
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
