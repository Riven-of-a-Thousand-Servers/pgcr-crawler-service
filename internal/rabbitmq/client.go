package rabbitmq

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Riven-of-a-Thousand-Servers/rivenbot-commons/pkg/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PgcrPublisher interface {
	Publish(pgcr types.PostGameCarnageReportResponse) error
}

type Rabbitmq struct {
	Channel *amqp.Channel
	Queue   *amqp.Queue
}

func NewRabbit(url, queueName string) (*Rabbitmq, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to RabbitMQ: %w", err)
	}
	channel, err := conn.Channel()
	if err != nil {
		return nil, errors.New("Error opening a channel")
	}

	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("Unable to declare a queue with name [%s]", queueName)
	}

	return &Rabbitmq{
		Channel: channel,
		Queue:   &queue,
	}, nil
}

// Publish a PGCR onto the queue for processing
func (p *Rabbitmq) Publish(pgcr types.PostGameCarnageReportResponse) error {
	body, err := json.Marshal(pgcr)
	if err != nil {
		return fmt.Errorf("Error marshalling pgcr")
	}

	err = p.Channel.Publish(
		"",
		p.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("There was an error publishing PGCR [%d]: %w", pgcr.Response.ActivityDetails.ReferenceId, err)
	}
	return nil
}
