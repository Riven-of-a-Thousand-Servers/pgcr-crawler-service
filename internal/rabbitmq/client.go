package rabbitmq

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Riven-of-a-Thousand-Servers/rivenbot-commons/pkg/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher interface {
	Publish(pgcr types.PostGameCarnageReport) error
}

type PgcrPublisher struct {
	Channel *amqp.Channel
	Queue   *amqp.Queue
}

// Publish a PGCR onto the queue for processing
func (p *PgcrPublisher) Publish(pgcr types.PostGameCarnageReport) error {
	body, err := json.Marshal(pgcr)
	if err != nil {
		return fmt.Errorf("Error marshalling pgcr")
	}

	p.Channel.Publish(
		"",
		p.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	return nil
}

func Connect() (*amqp.Connection, error) {
	connectionUrl := "amqp://localhost:5672"
	conn, err := amqp.Dial(connectionUrl)
	if err != nil {
		return nil, fmt.Errorf("Error opening up connection with url [%s]", connectionUrl)
	}
	return conn, nil
}

func Setup(connection *amqp.Connection, queueName string) (*amqp.Channel, *amqp.Queue, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, nil, errors.New("Error opening a channel")
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
		return nil, nil, fmt.Errorf("Unable to declare a queue with name [%s]", queueName)
	}

	return channel, &queue, nil
}
