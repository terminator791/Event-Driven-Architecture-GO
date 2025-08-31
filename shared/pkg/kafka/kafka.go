package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer defines the interface for publishing messages to Kafka
type ProducerInterface interface {
	Publish(ctx context.Context, key, value []byte) error
	Close() error
}

// Producer wraps kafka producer functionality
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new kafka producer
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// Publish sends a message to kafka
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	})
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Consumer wraps kafka consumer functionality
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new kafka consumer
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
	}
}

// Consume reads messages from kafka
func (c *Consumer) Consume(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// CreateTopic creates a kafka topic
func CreateTopic(brokers []string, topic string, partitions, replicationFactor int) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to dial controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Printf("Topic %s might already exist: %v", topic, err)
	}

	return nil
}