//go:build prod

package main

import (
	"log"
	"os"
	"pgcr-crawler-service/internal/bungie"
	"pgcr-crawler-service/internal/rabbitmq"
	"pgcr-crawler-service/internal/worker"
	"sync"

	"github.com/Riven-of-a-Thousand-Servers/rivenbot-commons/pkg/utils"
)

func run() {
	log.Printf(`
		Goroutines: %d
		Period: %d
		`)
	apiKeyFile := os.Getenv("BUNGIE_API_KEY_FILE")
	if apiKeyFile == "" {
		log.Fatal("Bungie api key not found")
	}

	apiKey, err := utils.ReadSecret(apiKeyFile)
	if err != nil {
		log.Fatalf("Error reading secret from file path [%s]", apiKeyFile)
	}

	hostPath := os.Getenv("BUNGIE_HOST_FILE")
	if hostPath == "" {
		log.Fatalf("Unable to host path for crawler")
	}

	host, err := utils.ReadSecret(hostPath)
	if err != nil {
		log.Fatalf("Unable to fetch secret with path [%s]", hostPath)
	}

	bungieClient, err := bungie.NewBungieClient(apiKey, host)
	if err != nil {
		log.Fatalf("Unable to instantiate a bungie client: %w", err)
	}

	rabbitmqPath := os.Getenv("RABBITMQ_URL_FILE")
	if rabbitmqPath == "" {
		log.Fatal("Rabbitmq url not found")
	}

	rabbitmqUrl, err := utils.ReadSecret(rabbitmqPath)
	if err != nil {
		log.Fatal("Error reading rabbitmq url from path: %s", rabbitmqUrl)
	}
	rabbitmq, err := rabbitmq.NewRabbit(rabbitmqUrl, "pgcr")
	if err != nil {
		log.Fatalf("Unable to instantiate rabbitmq: %w", err)
	}

	var waitgroup sync.WaitGroup
	ids := make(chan int64, 50)

	for {
		for i := 0; i <= *goroutines; i++ {
			waitgroup.Add(1)
			go func(wg *sync.WaitGroup, ids chan int64) {
				defer waitgroup.Done()
				worker := worker.Worker{
					BungieClient:    bungieClient,
					RabbitPublisher: rabbitmq,
				}

				for instanceId := range ids {
					worker.Work(instanceId)
				}
			}(&waitgroup, ids)
		}

		for i := 0; i < int(*period); i++ {
			ids <- int64(i)
		}

		close(ids)
		waitgroup.Wait()
	}
}
