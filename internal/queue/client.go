package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Host                string
	Port                string
	User                string
	Pass                string
	Exchange            string
	DeadLetterExchange  string
	UserRegisteredQueue string
}

type Client struct {
	cfg        Config
	conn       *amqp.Connection
	pubChannel *amqp.Channel
	confirms   <-chan amqp.Confirmation
	pubMu      sync.Mutex
}

func NewClient(cfg Config) (*Client, error) {
	conn, err := amqp.Dial(amqpURL(cfg))
	if err != nil {
		return nil, fmt.Errorf("connect RabbitMQ: %w", err)
	}

	pubChannel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open RabbitMQ publish channel: %w", err)
	}

	client := &Client{
		cfg:        cfg,
		conn:       conn,
		pubChannel: pubChannel,
	}
	if err := client.declareTopology(pubChannel); err != nil {
		_ = pubChannel.Close()
		_ = conn.Close()
		return nil, err
	}
	if err := pubChannel.Confirm(false); err != nil {
		_ = pubChannel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("enable RabbitMQ publisher confirms: %w", err)
	}
	client.confirms = pubChannel.NotifyPublish(make(chan amqp.Confirmation, 1))

	return client, nil
}

func (c *Client) Close() error {
	var errs []error
	if c.pubChannel != nil {
		errs = append(errs, c.pubChannel.Close())
	}
	if c.conn != nil {
		errs = append(errs, c.conn.Close())
	}

	return errors.Join(errs...)
}

func (c *Client) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	return c.conn.NotifyClose(receiver)
}

func (c *Client) PublishUserRegistered(ctx context.Context, event UserRegisteredEvent) error {
	return c.publish(ctx, UserRegisteredRoutingKey, event)
}

func (c *Client) ConsumeUserRegistered(ctx context.Context) (<-chan amqp.Delivery, error) {
	channel, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open RabbitMQ consume channel: %w", err)
	}

	if err := c.declareTopology(channel); err != nil {
		_ = channel.Close()
		return nil, err
	}
	if err := channel.Qos(1, 0, false); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("configure RabbitMQ qos: %w", err)
	}

	deliveries, err := channel.ConsumeWithContext(
		ctx,
		c.cfg.UserRegisteredQueue,
		"web-labs-user-registered-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("start RabbitMQ consumer: %w", err)
	}

	go func() {
		<-ctx.Done()
		_ = channel.Close()
	}()

	return deliveries, nil
}

func (c *Client) publish(ctx context.Context, routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal RabbitMQ message: %w", err)
	}

	c.pubMu.Lock()
	defer c.pubMu.Unlock()

	err = c.pubChannel.PublishWithContext(
		ctx,
		c.cfg.Exchange,
		routingKey,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish RabbitMQ message: %w", err)
	}

	select {
	case confirmation, ok := <-c.confirms:
		if !ok {
			return errors.New("RabbitMQ publisher confirm channel closed")
		}
		if !confirmation.Ack {
			return errors.New("RabbitMQ broker rejected published message")
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("wait RabbitMQ publisher confirm: %w", ctx.Err())
	case <-time.After(5 * time.Second):
		return errors.New("timed out waiting for RabbitMQ publisher confirm")
	}
}

func (c *Client) declareTopology(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(
		c.cfg.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare RabbitMQ exchange: %w", err)
	}

	if err := channel.ExchangeDeclare(
		c.cfg.DeadLetterExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare RabbitMQ dead letter exchange: %w", err)
	}

	_, err := channel.QueueDeclare(
		c.cfg.UserRegisteredQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    c.cfg.DeadLetterExchange,
			"x-dead-letter-routing-key": UserRegisteredRoutingKey,
		},
	)
	if err != nil {
		return fmt.Errorf("declare RabbitMQ user registered queue: %w", err)
	}

	if err := channel.QueueBind(
		c.cfg.UserRegisteredQueue,
		UserRegisteredRoutingKey,
		c.cfg.Exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind RabbitMQ user registered queue: %w", err)
	}

	_, err = channel.QueueDeclare(
		c.deadLetterQueueName(),
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare RabbitMQ user registered dlq: %w", err)
	}

	if err := channel.QueueBind(
		c.deadLetterQueueName(),
		UserRegisteredRoutingKey,
		c.cfg.DeadLetterExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind RabbitMQ user registered dlq: %w", err)
	}

	return nil
}

func (c *Client) deadLetterQueueName() string {
	return c.cfg.UserRegisteredQueue + ".dlq"
}

func amqpURL(cfg Config) string {
	user := url.UserPassword(cfg.User, cfg.Pass)
	return fmt.Sprintf("amqp://%s@%s/", user.String(), net.JoinHostPort(cfg.Host, cfg.Port))
}
