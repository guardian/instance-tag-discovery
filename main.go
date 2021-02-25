package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Prune EBS snapshots - example script using AWS SDK.
func main() {

	// 1. Load credentials/config
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("eu-west-1"),
		config.WithSharedConfigProfile("frontend"),
	)
	check(err, "Error loading config")

	// 2. Initialise client for service you want (here ec2 is used)
	client := ec2.NewFromConfig(cfg)

	// 3. Build and execute requests
	input := ec2.DescribeSnapshotsInput{}
	output, err := client.DescribeSnapshots(context.TODO(), &input)
	check(err, "Describe snapshots failed")

	// (4. Do some extra stuff)
	sixtyDaysAgo := time.Now().AddDate(0, 0, -60)

	fmt.Println("Deleting snapshots...")
	for _, s := range output.Snapshots {
		if s.StartTime.Before(sixtyDaysAgo) && *s.OwnerId == "SOME ID" {
			fmt.Printf("%s %s\n", s.StartTime.Format(time.RFC3339), *s.SnapshotId)
			_, err = client.DeleteSnapshot(context.TODO(), &ec2.DeleteSnapshotInput{SnapshotId: s.SnapshotId, DryRun: false})
			check(err, fmt.Sprintf("Unable to delete snapshot %s", *s.SnapshotId))
		}
	}
}

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s; %v", msg, err)
	}
}
