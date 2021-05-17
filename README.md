# Go AWS example

    brew install go

Then check your version using `go version`. You'll need 1.15 or greater to use
the AWS V2 SDK.

Go has in-built dependency management using [Go
mod](https://blog.golang.org/using-go-modules).

The AWS SDK for Go has docs here:

- https://aws.github.io/aws-sdk-go-v2/docs/ (developer guide)
- https://pkg.go.dev/github.com/aws/aws-sdk-go-v2 (API reference)

# Deployment

This tool is baked into images in the cdk-base role in AMIgo. In order to update it you should grab the built artifacts from the GitHub actions build and upload them to `packages/deb/` in the AMIgo data bucket.

It would be nice to make this work with RiffRaff at some stage.