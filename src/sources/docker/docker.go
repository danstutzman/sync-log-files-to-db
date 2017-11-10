package docker

import (
	"context"
	"log"

	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
	"github.com/docker/docker/api/types"
	"github.com/moby/moby/client"
)

type Options struct {
	Influxdb *influxdb.Options
}

func Main() {
	client, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Error from NewEnvClient: %s", err)
	}

	containers, err := client.ContainerList(
		context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatalf("Error from ContainerList: %s", err)
	}

	for _, container := range containers {
		log.Printf("Tailing logs for container %s (image=%s)...",
			container.ID[:10], container.Image)

		out, err := client.ContainerLogs(
			context.Background(),
			container.ID,
			types.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				// Since      string
				Timestamps: true,
				Follow:     true,
				// Tail:       string, // num lines from end to show?
				Details: true,
			})
		if err != nil {
			panic(err)
		}

		for {
			logLine := readLogLineBlocking(out)
			log.Printf("%v", logLine)
		}
	}
}
