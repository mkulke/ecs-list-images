package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"os"
)

var cluster string

func listTasks(svc *ecs.ECS) ([]*string, error) {
	params := &ecs.ListTasksInput{
		Cluster:    aws.String("default"),
		MaxResults: aws.Int64(64),
	}

	resp, err := svc.ListTasks(params)
	if err != nil {
		return nil, err
	}

	return resp.TaskArns, nil
}

func getImages(svc *ecs.ECS, definitions []*string) ([]string, error) {
	images := []string{}
	for _, definition := range definitions {
		params := &ecs.DescribeTaskDefinitionInput{
			TaskDefinition: definition,
		}
		resp, err := svc.DescribeTaskDefinition(params)
		if err != nil {
			return nil, err
		}
		containers := resp.TaskDefinition.ContainerDefinitions
		for _, container := range containers {
			images = append(images, *container.Image)
		}
	}

	return images, nil
}

func getTaskDefinitions(svc *ecs.ECS, tasks []*string) ([]*string, error) {
	params := &ecs.DescribeTasksInput{
		Cluster: aws.String("default"),
		Tasks:   tasks,
	}
	resp, err := svc.DescribeTasks(params)
	if err != nil {
		return nil, err
	}

	set := make(map[*string]struct{})
	for _, task := range resp.Tasks {
		set[task.TaskDefinitionArn] = struct{}{}
	}

	definitions := []*string{}
	for definition := range set {
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

type Output struct {
	Images []string `json:"images"`
}

func main() {
	flag.StringVar(&cluster, "cluster", os.Getenv("ECS_CLUSTER"), "ecs cluster")
	flag.Parse()

	if cluster == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	svc := ecs.New(session.New())

	tasks, err := listTasks(svc)
	if err != nil {
		panic(err)
	}

	definitions, err := getTaskDefinitions(svc, tasks)
	if err != nil {
		panic(err)
	}

	images, err := getImages(svc, definitions)
	if err != nil {
		panic(err)
	}

	output := Output{images}
	bytes, err := json.Marshal(output)
	if err != nil {
		panic(err)
	}
	fmt.Printf(string(bytes))
}
