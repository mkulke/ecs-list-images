ECS-List-Images
===

About
==

A simple tool to list deployed docker images in ECS clusters. It scans a specified cluster and outputs a list of found images as json. Caveat: clusters with > 64 tasks, and environments with > 64 clusters are not supported, yet.

Usage
==

Using the AWS api library, auth is handled the usual way via config in `~/.aws/`, environment variables or metadata on EC2 instances, refer to the AWS cli documentation for specifics.

Arguments for the tool can be substituted with env variables (arguments prefixed by `ECS_`).

A cluster can be specified using its name `--cluster mycluster`, otherwise the images from all clusters are collected.

```
export AWS_REGION=eu-central-1
./ecs-list-images -cluster mycluster
ECS_CLUSTER=default ./ecs-list-images
```

Build
==

Go 1.5+ on OSX targetting linux:

```
go get -u github.com/aws/aws-sdk-go
GOOS=linux go build ecs-list-images
```
