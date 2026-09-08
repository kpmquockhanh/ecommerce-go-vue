package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"ecommerce-api-go/internal/repositories"
)

const MaxRetries = 3

type JobType string

const (
	JobOrderConfirmation  JobType = "order_confirmation"
	JobPaymentFailed      JobType = "payment_failed"
	JobOrderStatusChanged JobType = "order_status_changed"
)

type Job struct {
	Type       JobType     `json:"type"`
	Payload    interface{} `json:"payload"`
	CreatedAt  time.Time   `json:"created_at"`
	RetryCount int         `json:"retry_count"`
}

type OrderConfirmationPayload struct {
	OrderID       int    `json:"order_id"`
	UserID        int    `json:"user_id"`
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Total         int    `json:"total"`
	PaymentIntent string `json:"payment_intent"`
}

type PaymentFailedPayload struct {
	OrderID   int    `json:"order_id"`
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Total     int    `json:"total"`
	Reason    string `json:"reason"`
}

type OrderStatusChangedPayload struct {
	OrderID   int    `json:"order_id"`
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

type Queue struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	deadLetterRepo repositories.DeadLetterRepository
	uri            string
}

func New(rabbitURI string, deadLetterRepo repositories.DeadLetterRepository) (*Queue, error) {
	q := &Queue{uri: rabbitURI, deadLetterRepo: deadLetterRepo}
	if err := q.connect(); err != nil {
		return nil, err
	}
	return q, nil
}

func (q *Queue) connect() error {
	var err error
	q.conn, err = amqp.Dial(q.uri)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	q.channel, err = q.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare main queues
	queues := []string{"emails", "jobs"}
	for _, name := range queues {
		_, err := q.channel.QueueDeclare(
			name,  // name
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", name, err)
		}
	}

	// Declare DLQ
	_, err = q.channel.QueueDeclare(
		"dead_letters",
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": "dead_letters",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter queue: %w", err)
	}

	log.Println("Connected to RabbitMQ")
	return nil
}

func (q *Queue) Publish(ctx context.Context, queueName string, job Job) error {
	job.CreatedAt = time.Now()

	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	return q.channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
}

func (q *Queue) Consume(ctx context.Context, queueName string, handler func(Job) error) error {
	msgs, err := q.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // autoAck
		false,     // exclusive
		false,     // noLocal
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("consumer channel closed")
			}

			var job Job
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				log.Printf("Failed to unmarshal job: %v", err)
				q.persistDeadLetter(ctx, queueName, job, err.Error())
				msg.Nack(false, false)
				continue
			}

			if err := handler(job); err != nil {
				log.Printf("Failed to process job %s (attempt %d/%d): %v",
					job.Type, job.RetryCount+1, MaxRetries, err)

				if job.RetryCount < MaxRetries {
					// Requeue with incremented retry count
					job.RetryCount++
					if pubErr := q.Publish(ctx, queueName, job); pubErr != nil {
						log.Printf("Failed to republish job: %v", pubErr)
					}
					msg.Ack(false)
				} else {
					// Max retries exceeded — persist to DLQ and ack the original message
					q.persistDeadLetter(ctx, queueName, job, err.Error())
					msg.Ack(false)
				}
			} else {
				msg.Ack(false)
			}
		}
	}
}

func (q *Queue) persistDeadLetter(ctx context.Context, queueName string, job Job, errMsg string) {
	if q.deadLetterRepo == nil {
		log.Printf("DLQ: no dead letter repository, job lost: type=%s queue=%s", job.Type, queueName)
		return
	}

	payload, err := json.Marshal(job.Payload)
	if err != nil {
		payload = []byte("{}")
	}

	if err := q.deadLetterRepo.Persist(ctx, queueName, string(job.Type), payload, errMsg, job.RetryCount); err != nil {
		log.Printf("DLQ: failed to persist dead letter: %v (job: %s)", err, job.Type)
	} else {
		log.Printf("DLQ: persisted failed job type=%s queue=%s retries=%d", job.Type, queueName, job.RetryCount)
	}
}

func (q *Queue) Close() error {
	if q.channel != nil {
		q.channel.Close()
	}
	if q.conn != nil {
		return q.conn.Close()
	}
	return nil
}
