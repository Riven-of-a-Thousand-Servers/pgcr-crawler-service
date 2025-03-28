//go:build prod

package main

import (
	"github.com/Riven-of-a-Thousand-Servers/rivenbot-commons/pkg/utils"
	"log"
	"net/http"
	"os"
	"pgcr-crawler-service/internal/bungie"
	"pgcr-crawler-service/internal/rabbitmq"
	"pgcr-crawler-service/internal/worker"
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

	httpClient := &http.Client{}

	bungieClient := &bungie.PgcrClient{
		Client: httpClient,
		Host:   host,
		ApiKey: apiKey,
	}

	conn, err := rabbitmq.Connect()
	if err != nil {
		log.Fatalf("%v", err)
	}

	defer conn.Close()
	channel, queue, err := rabbitmq.Setup(conn, "pgcr-crawler-service")
	if err != nil {
		log.Fatalf("%v", err)
	}

	rabbitPublisher := &rabbitmq.PgcrPublisher{
		Channel: channel,
		Queue:   queue,
	}

	worker := worker.Worker{
		BungieClient:    bungieClient,
		RabbitPublisher: rabbitPublisher,
	}

	worker.Work(1)

}
