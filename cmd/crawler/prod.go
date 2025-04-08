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
	"github.com/rabbitmq/amqp091-go"
)

func run() {
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
		log.Fatalf("Error reading rabbitmq url from path [%s]: %v", rabbitmqUrl, err)
	}

	conn, err := amqp091.Dial(rabbitmqUrl)
	if err != nil {
		log.Fatalf("Error connecting to rabbitmq: %v", err)
	}

	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error opening up channel: %v", err)
	}

	defer channel.Close()

	rabbitmq, err := rabbitmq.NewRabbit(rabbitmqUrl, "pgcr", channel)
	if err != nil {
		log.Fatalf("Unable to instantiate rabbitmq: %w", err)
	}

	var waitgroup sync.WaitGroup

	for i := 0; i < *passes; i++ {
		ids := make(chan int64, 50)
		for i := 0; i <= *goroutines; i++ {
			waitgroup.Add(1)
			go func(wg *sync.WaitGroup, ids chan int64) {
				defer waitgroup.Done()
				worker := worker.NewWorker(bungieClient, rabbitmq)

				for instanceId := range ids {
					id := instanceId
					if *pgcrStartingPoint != -1 {
						id = *pgcrStartingPoint + instanceId
					}
					log.Printf("Working on pgcr [%d]...", id)
					worker.Work(id)
					log.Printf("Finished pgcr [%d]", id)
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
