package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	cli               *client.Client
	errInvalidNetwork = errors.New("invalid network")
)

type networkHealthCheck struct {
	name          string
	status        string
	failingStreak float64
}

// The docker API accepts the network IDs to inspect, but we can use the names of the containers to make them friendly / memorable.

func init() {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	cli = client
}

// returns the name of networks and their corresponding ID's.
// Eg. bridge -> 122121233eeed
func ListNetworks() (map[string]string, error) {
	networkList, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}
	networkDetails := make(map[string]string, len(networkList))
	for _, network := range networkList {
		networkDetails[network.Name] = network.ID
	}
	return networkDetails, nil
}

// Pass the network name into this function, and it returns the health status of
// all the containers in this network.
func InspectNetworkByName(name string) ([]networkHealthCheck, error) {
	networks, err := ListNetworks()
	if err != nil {
		return nil, err
	}
	found := false
	networkID := ""

	for networkName, ID := range networks {
		if networkName == name {
			found = true
			networkID = ID
			break
		}
	}

	if !found {
		return nil, errInvalidNetwork
	}

	resource, err := cli.NetworkInspect(context.Background(), networkID, types.NetworkInspectOptions{})
	if err != nil {
		return nil, err
	}
	networkHealthChecks := make([]networkHealthCheck, 0, len(resource.Containers))
	// fetch containers in this network
	for ID := range resource.Containers {
		containerDetails, raw, err := cli.ContainerInspectWithRaw(context.Background(), ID, true)
		if err != nil {
			return nil, err
		}

		healthCheck := containerDetails.Config.Healthcheck
		record := networkHealthCheck{name: containerDetails.Config.Labels["com.docker.compose.service"]}
		// Ensure it has a health check registered
		if healthCheck != nil {
			var response map[string]interface{}
			if err := json.Unmarshal(raw, &response); err != nil {
				return nil, err
			}
			health := response["State"].(map[string]interface{})["Health"].(map[string]interface{})
			record.status = health["Status"].(string)
			record.failingStreak = health["FailingStreak"].(float64)
		} else {
			record.status = "nil"
			record.failingStreak = 0
		}
		networkHealthChecks = append(networkHealthChecks, record)
	}
	return networkHealthChecks, nil
}
