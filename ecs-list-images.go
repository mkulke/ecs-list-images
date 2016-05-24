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

const maxResults = 64

var selectedCluster string

func listClusters(svc *ecs.ECS) ([]*string, error) {
	params := &ecs.ListClustersInput{
		MaxResults: aws.Int64(maxResults),
	}

	resp, err := svc.ListClusters(params)
	if err != nil {
		return nil, err
	}

	return resp.ClusterArns, nil
}

func listTasks(svc *ecs.ECS, cluster string) ([]*string, error) {
	params := &ecs.ListTasksInput{
		Cluster:    aws.String(cluster),
		MaxResults: aws.Int64(maxResults),
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

func getTaskDefinitions(svc *ecs.ECS, cluster string, tasks []*string) ([]*string, error) {
  if len(tasks) == 0 {
    return []*string{}, nil
  }

	params := &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
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

func getImagesForCluster(svc *ecs.ECS, cluster string) ([]string, error) {
	tasks, err := listTasks(svc, cluster)
	if err != nil {
		return nil, err
	}

	definitions, err := getTaskDefinitions(svc, cluster, tasks)
	if err != nil {
		return nil, err
	}

	images, err := getImages(svc, definitions)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func makeUnique(arr []string) []string {
	set := make(map[string]struct{})
	for _, entry := range arr {
		set[entry] = struct{}{}
	}
	uniqueArr := []string{}
	for entry := range set {
		uniqueArr = append(uniqueArr, entry)
	}
	return uniqueArr
}

func containsCluster(clusters []*string, cluster string) bool {
	for _, _cluster := range clusters {
		if cluster == *_cluster {
			return true
		}
	}
	return false
}

type output struct {
	Images []string `json:"images"`
}

func main() {
	flag.StringVar(&selectedCluster, "cluster", os.Getenv("ECS_CLUSTER"), "ecs cluster")
	flag.Parse()

	svc := ecs.New(session.New())

	clusters, err := listClusters(svc)
	if err != nil {
		panic(err)
	}

	if selectedCluster != "" {
		if !containsCluster(clusters, selectedCluster) {
			fmt.Printf("Cluster %s not found.\n", selectedCluster)
			os.Exit(1)
		}
		clusters = []*string{&selectedCluster}
	}

	images := []string{}
	for _, cluster := range clusters {
		clusterImages, err := getImagesForCluster(svc, *cluster)
		if err != nil {
			panic(err)
		}
		images = append(images, clusterImages...)
	}
	images = makeUnique(images)

	output := output{images}
	bytes, err := json.Marshal(output)
	if err != nil {
		panic(err)
	}
	fmt.Printf(string(bytes))
}
