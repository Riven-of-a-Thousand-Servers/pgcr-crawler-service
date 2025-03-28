package worker

import (
	"pgcr-crawler-service/internal/bungie"
	"pgcr-crawler-service/internal/rabbitmq"
)

type Worker struct {
	BungieClient    bungie.BungieClient
	RabbitPublisher rabbitmq.RabbitPublisher
}

func (w Worker) Work(instanceId int64) {
	return
}
