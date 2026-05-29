package queue

import (
	"context"
	"encoding/json"
	"log"

	"Web_lab/internal/mailer"

	amqp "github.com/rabbitmq/amqp091-go"
)

const maxDeliveryAttempts = 3

type Consumer struct {
	client *Client
	mailer *mailer.Mailer
}

func NewConsumer(client *Client, mailer *mailer.Mailer) *Consumer {
	return &Consumer{
		client: client,
		mailer: mailer,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.client.ConsumeUserRegistered(ctx)
	if err != nil {
		return err
	}

	go c.handleDeliveries(ctx, deliveries)
	return nil
}

func (c *Consumer) handleDeliveries(ctx context.Context, deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}
			c.handleUserRegistered(ctx, delivery)
		}
	}
}

func (c *Consumer) handleUserRegistered(ctx context.Context, delivery amqp.Delivery) {
	var event UserRegisteredEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		log.Printf("failed to decode user registered event: %v", err)
		c.reject(delivery)
		return
	}

	if event.Metadata.Attempt <= 0 {
		event.Metadata.Attempt = 1
	}

	if err := c.mailer.SendWelcomeEmail(event.Payload.Email); err != nil {
		log.Printf(
			"failed to send welcome email to %s on attempt %d: %v",
			event.Payload.Email,
			event.Metadata.Attempt,
			err,
		)
		c.retryOrDeadLetter(ctx, delivery, event)
		return
	}

	if err := delivery.Ack(false); err != nil {
		log.Printf("failed to ack user registered event %s: %v", event.EventID, err)
		return
	}
	log.Printf("welcome email sent to %s for event %s", event.Payload.Email, event.EventID)
}

func (c *Consumer) retryOrDeadLetter(ctx context.Context, delivery amqp.Delivery, event UserRegisteredEvent) {
	if event.Metadata.Attempt >= maxDeliveryAttempts {
		if err := delivery.Nack(false, false); err != nil {
			log.Printf("failed to dead-letter user registered event %s: %v", event.EventID, err)
		}
		return
	}

	event.Metadata.Attempt++
	if err := c.client.PublishUserRegistered(ctx, event); err != nil {
		log.Printf("failed to republish user registered event %s: %v", event.EventID, err)
		if nackErr := delivery.Nack(false, true); nackErr != nil {
			log.Printf("failed to requeue user registered event %s: %v", event.EventID, nackErr)
		}
		return
	}

	if err := delivery.Ack(false); err != nil {
		log.Printf("failed to ack retried user registered event %s: %v", event.EventID, err)
	}
}

func (c *Consumer) reject(delivery amqp.Delivery) {
	if err := delivery.Nack(false, false); err != nil {
		log.Printf("failed to reject invalid user registered event: %v", err)
	}
}
