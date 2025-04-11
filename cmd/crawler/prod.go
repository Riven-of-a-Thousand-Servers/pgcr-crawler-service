//go:build prod

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"pgcr-crawler-service/internal/bungie"
	"pgcr-crawler-service/internal/rabbitmq"
	"pgcr-crawler-service/internal/worker"
	"sync"
	"time"

	"github.com/Riven-of-a-Thousand-Servers/rivenbot-commons/pkg/utils"
	"github.com/rabbitmq/amqp091-go"
)

func checkProxyHealth(proxyHost string) error {
	url := fmt.Sprintf("http://%s/healthcheck", proxyHost)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Failed to reach proxy healthcheck endpoint: %v", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Healthcheck failed with status: %s", resp.Status)
	}

	return nil
}

func run() {
	apiKey, host, rabbitmqUrl := parseEnv()
	bungieClient, err := bungie.NewBungieClient(apiKey, host)
	if err != nil {
		log.Fatalf("Unable to instantiate a bungie client: %v", err)
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
		log.Fatalf("Unable to instantiate rabbitmq: %v", err)
	}

	crawl(bungieClient, rabbitmq, host)
}

func crawl(bungieClient bungie.PgcrClient, rabbitmq rabbitmq.PgcrPublisher, proxyHost string) {
	waitForHealthyProxy(proxyHost)
	var waitgroup sync.WaitGroup
	for i := 0; i < *passes; i++ {
		ids := make(chan int64, 5)
		for i := 0; i <= *goroutines; i++ {
			waitgroup.Add(1)
			go func(ids chan int64) {
				defer waitgroup.Done()
				worker := worker.NewWorker(bungieClient, rabbitmq)

				for instanceId := range ids {
					id := instanceId
					if *pgcrStartingPoint != -1 {
						id = *pgcrStartingPoint + instanceId
					}
					log.Printf("Working on pgcr [%d]...", id)
					err := worker.Work(id)
					if err != nil {
						log.Printf("Received error for pgcr [%d]: %v", id, err)
					}
					log.Printf("Finished pgcr [%d]", id)
				}
			}(ids)
		}

		for i := 0; i < int(*period); i++ {
			ids <- int64(i)
		}

		close(ids)
		waitgroup.Wait()
	}
}

func waitForHealthyProxy(proxyHost string) {
	healthy := false
	maxAttempts := 5
	for i := 0; i < maxAttempts; i++ {
		err := checkProxyHealth(proxyHost)
		if err != nil {
			log.Printf("Attempt %d/%d: Proxy health check failed : %v", i+1, maxAttempts, err)
			time.Sleep(15 * time.Second)
			continue
		}
		log.Println("Proxy is healthy, starting the crawler...")
		healthy = true
		break
	}

	if !healthy {
		log.Fatalf("Proxy did not become healthy after 5 attemps. Exiting")
	}
}

func parseEnv() (apiKey, proxyHost, rabbitmqUrl string) {
	apiKeyFile := os.Getenv("BUNGIE_API_KEY_FILE")
	if apiKeyFile == "" {
		log.Fatal("Bungie api key not found")
	}

	apiKey, err := utils.ReadSecret(apiKeyFile)
	if err != nil {
		log.Fatalf("Error reading secret from file path [%s]: %v", apiKeyFile, err)
	}

	hostPath := os.Getenv("BUNGIE_HOST_FILE")
	if hostPath == "" {
		log.Fatalf("Unable to host path for crawler")
	}

	proxyHost, err = utils.ReadSecret(hostPath)
	if err != nil {
		log.Fatalf("Unable to fetch secret with path [%s]: %v", hostPath, err)
	}

	rabbitmqPath := os.Getenv("RABBITMQ_URL_FILE")
	if rabbitmqPath == "" {
		log.Fatal("Rabbitmq url not found")
	}

	rabbitmqUrl, err = utils.ReadSecret(rabbitmqPath)
	if err != nil {
		log.Fatalf("Error reading rabbitmq url from path [%s]: %v", rabbitmqUrl, err)
	}

	return apiKey, proxyHost, rabbitmqUrl
}
