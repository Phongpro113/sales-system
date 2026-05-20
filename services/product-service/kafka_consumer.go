package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
)

const topicOrderCreated = "order.created"

type OrderCreatedEvent struct {
	OrderID uint        `json:"order_id"`
	Items   []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type stockConsumer struct{}

func (c *stockConsumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (c *stockConsumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (c *stockConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal order.created: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		log.Printf("Consumed order.created: orderID=%d, items=%d", event.OrderID, len(event.Items))
		processStockReduction(event)
		session.MarkMessage(msg, "")
	}
	return nil
}

func processStockReduction(event OrderCreatedEvent) {
	for _, item := range event.Items {
		var product Product
		if err := db.First(&product, item.ProductID).Error; err != nil {
			log.Printf("Product %d not found for stock reduction: %v", item.ProductID, err)
			continue
		}

		newStock := product.Stock - item.Quantity
		if newStock < 0 {
			log.Printf("Warning: stock underflow for product %d (stock=%d, qty=%d)", item.ProductID, product.Stock, item.Quantity)
			newStock = 0
		}

		if err := db.Model(&product).Update("stock", newStock).Error; err != nil {
			log.Printf("Failed to reduce stock for product %d: %v", item.ProductID, err)
		} else {
			log.Printf("Stock reduced: product %d, %d -> %d", item.ProductID, product.Stock, newStock)
		}
	}
}

func startKafkaConsumer() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	cfg := sarama.NewConfig()
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	var group sarama.ConsumerGroup
	var err error
	for i := 0; i < 10; i++ {
		group, err = sarama.NewConsumerGroup([]string{brokers}, "product-service", cfg)
		if err == nil {
			break
		}
		log.Printf("Kafka not ready, retrying (%d/10): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer group: %v", err)
	}

	log.Println("Kafka consumer started, listening on topic:", topicOrderCreated)

	go func() {
		defer group.Close()
		handler := &stockConsumer{}
		for {
			if err := group.Consume(context.Background(), []string{topicOrderCreated}, handler); err != nil {
				log.Printf("Kafka consumer error: %v", err)
				time.Sleep(2 * time.Second)
			}
		}
	}()
}
