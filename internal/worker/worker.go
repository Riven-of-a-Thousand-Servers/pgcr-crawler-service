package worker

import (
	"fmt"
	"pgcr-crawler-service/internal/bungie"
	"pgcr-crawler-service/internal/rabbitmq"
)

type Worker struct {
	BungieClient    bungie.PgcrClient
	RabbitPublisher rabbitmq.PgcrPublisher
}

func NewWorker(bungieClient bungie.PgcrClient, rabbitPublisher rabbitmq.PgcrPublisher) *Worker {
	return &Worker{
		BungieClient:    bungieClient,
		RabbitPublisher: rabbitPublisher,
	}
}

func (w *Worker) Work(instanceId int64) error {
	pgcr, err := w.BungieClient.FetchPgcr(instanceId)
	if err != nil {
		return fmt.Errorf("Error fetching instanceId [%d] from Bungie: %v", instanceId, err)
	}

	err = w.RabbitPublisher.Publish(pgcr.Response)
	if err != nil {
		return fmt.Errorf("Error publishing instanceId [%d] to RabbitMQ: %v", instanceId, err)
	}
	return nil
}
