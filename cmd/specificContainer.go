package cmd

import (
    "context"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/filters"
    dockerClient "github.com/docker/docker/client"
)

var containerName string

func init() {
    rootCmd.Flags().StringVarP(&containerName, "container", "c", "", "指定容器名")
}

func filterContainer(containerName string) types.Container {
    // list container
    cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
    if err != nil {
        panic(err)
    }
    defer cli.Close()
    containers, err := cli.ContainerList(context.Background(), container.ListOptions{Filters: filters.NewArgs(filters.Arg("name", containerName))})
    if err != nil {
        panic(err)
    }
    if len(containers) == 0 {
        panic("未找到指定容器: " + containerName)
    }
    return containers[0]
}
