# Instance Tag Discovery

A simple program to fetch instance tags and write them to disk.

Tags are fetched from (in order of preference):

* instance metadata (IMDS v2 - enable tags via [Launch Template
  metadata](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-launchtemplate-launchtemplatedata-metadataoptions.html#cfn-ec2-launchtemplate-launchtemplatedata-metadataoptions-instancemetadatatags))
* the EC2 instance (EC2 API)

They are written to:

* `/etc/config/tags.properties` (key=value format)
* `/etc/config/tags.json` (JSON format)

The `out-dir` flag can be used to change the target directory.

**We strongly recommend that you configure your instances to support tags on
their metadata. This will make lookup quicker and also avoid the possibility of
hitting rate limiting on the AWS API.**

## Local development

Install Go:

    brew install go

Then run some commands:

    go test
    go run main.go --profile [profile] --out-dir . --instance-id [some-id]

If unsure about editors, we recommend using VS Code.

Go has in-built dependency management using [Go
mod](https://blog.golang.org/using-go-modules).

The AWS SDK for Go has docs here:

- https://aws.github.io/aws-sdk-go-v2/docs/ (developer guide)
- https://pkg.go.dev/github.com/aws/aws-sdk-go-v2 (API reference)

# Deployment

This tool is baked into images in the cdk-base role in AMIgo. In order to update
it you should grab the built artifacts from the GitHub actions build and upload
them to `packages/instance-tag-discovery/` in the AMIgo data bucket.

It would be nice to make this work with RiffRaff at some stage.
