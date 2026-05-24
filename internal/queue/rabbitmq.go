package queue

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	JobExchange      = "jobs.exchange"
	JobQueue         = "jobs.execution"
	JobRoutingKey    = "job.new"

	RetryExchange    = "jobs.retry_exchange"
	RetryQueue       = "jobs.retry"
	RetryRoutingKey  = "job.retry"

	DLQQueue         = "jobs.dlq"
)

type RabbitClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitClient(amqpURL string) (*RabbitClient, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	client := &RabbitClient{conn: conn, channel: ch}
	if err := client.setupTopology(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

func (rc *RabbitClient) setupTopology() error {
	// Core Job Exchange
	if err := rc.channel.ExchangeDeclare(JobExchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// Retry Exchange (Handles delayed routing via TTL)
	if err := rc.channel.ExchangeDeclare(RetryExchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// Main Execution Queue
	_, err := rc.channel.QueueDeclare(JobQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	if err := rc.channel.QueueBind(JobQueue, JobRoutingKey, JobExchange, false, nil); err != nil {
		return err
	}

	// Delayed Retry Queue
	// configure this queue to automatically drop expired messages back into the main exchange
	_, err = rc.channel.QueueDeclare(
		RetryQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    JobExchange,    // Where to send when message expires
			"x-dead-letter-routing-key": JobRoutingKey, // Route back to core execution
		},
	)
	if err != nil {
		return err
	}
	if err := rc.channel.QueueBind(RetryQueue, RetryRoutingKey, RetryExchange, false, nil); err != nil {
		return err
	}

	// Dead Letter Queue (Final resting place for permanently failed jobs)
	_, err = rc.channel.QueueDeclare(DLQQueue, true, false, false, false, nil)
	return err
}

func (rc *RabbitClient) PublishJob(ctx context.Context, jobID string, taskName string, payload map[string]interface{}) error {
	body, err := json.Marshal(map[string]interface{}{"id": jobID, "name": taskName, "payload": payload})
	if err != nil {
		return err
	}

	return rc.channel.PublishWithContext(ctx, JobExchange, JobRoutingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    jobID,
			Body:         body,
		},
	)
}

// PublishToRetry queues the message with a specific time-to-live delay
func (rc *RabbitClient) PublishToRetry(ctx context.Context, body []byte, delayMs int) error {
	return rc.channel.PublishWithContext(ctx,
		RetryExchange,
		RetryRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Expiration:   fmt.Sprintf("%d", delayMs), // Message sits in retry queue for exactly this long
			Body:         body,
		},
	)
}

// PublishToDLQ moves the job off the retry path and sets it aside for analysis
func (rc *RabbitClient) PublishToDLQ(ctx context.Context, body []byte) error {
	return rc.channel.PublishWithContext(ctx,
		"", // Default exchange routes directly to a queue using its name as routing key
		DLQQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (rc *RabbitClient) Consume() (<-chan amqp.Delivery, error) {
	return rc.channel.Consume(JobQueue, "worker-instance", false, false, false, false, nil)
}

func (rc *RabbitClient) Close() {
	if rc.channel != nil { rc.channel.Close() }
	if rc.conn != nil { rc.conn.Close() }
}
