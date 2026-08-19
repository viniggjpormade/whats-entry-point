package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar env .env file: %s", err)
	}
	kafkaServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if kafkaServers == "" {
		kafkaServers = "localhost:9092"
	}

	config := &kafka.ConfigMap{
		"bootstrap.servers": kafkaServers,
		"acks":              "1",
		// "enable.idempotence":                    true,
		"retries":                               2147483647,
		"reconnect.backoff.ms":                  30000,
		"max.in.flight.requests.per.connection": 1,
		"linger.ms":                             5,
		"batch.num.messages":                    10000,
		"reconnect.backoff.max.ms":              30000,
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Falha ao criar produtor: %v", err)
	}
	defer producer.Close()
	go producerEventLoop(producer)

	http.HandleFunc("/whats/webhook", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			webhookVerify(w, r)
		case http.MethodPost:
			webhookHandle(producer, w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Servidor rodando na porta %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Encerrando o servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Servidor forçado a encerrar: %v", err)
	}

	log.Println("Fazendo flush no Kafka...")
	producer.Flush(15000)

	log.Println("Servidor encerrado com sucesso.")
}

func webhookHandle(producer *kafka.Producer, w http.ResponseWriter, r *http.Request) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("Webhook received at %s", timestamp)

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Erro ao ler o corpo da requisição: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Payload bruto recebido: %s", string(rawBody))

	var body WaReceivedResponse
	if err := json.Unmarshal(rawBody, &body); err != nil {
		log.Printf("Erro ao fazer parse do webhook: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid payload"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "received"}`))

	kafkaTopic := os.Getenv("KAFKA_TOPIC_WA")

	messageBytes, err := json.Marshal(body)
	if err != nil {
		log.Printf("Erro ao serializar mensagem para o Kafka: %v", err)
		return
	}

	err = sendMessage(producer, kafkaTopic, uuid.New().String(), messageBytes)
	if err != nil {
		log.Printf("Erro ao enfileirar no Kafka: %v", err)
	}
}

func webhookVerify(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	challenge := r.URL.Query().Get("hub.challenge")
	token := r.URL.Query().Get("hub.verify_token")
	verifyToken := os.Getenv("VERIFY_TOKEN")
	if mode == "subscribe" && token == verifyToken {
		log.Println("WEBHOOK VERIFIED")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}
	http.Error(w, "Forbidden", http.StatusForbidden)

}

func producerEventLoop(producer *kafka.Producer) {
	defer log.Println("")

	for event := range producer.Events() {
		switch ev := event.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				log.Printf("Falha ao entregar mensagem: %v", ev.TopicPartition.Error)
			} else {
				log.Printf("Mensagem entregue em %s[%d]@%d",
					*ev.TopicPartition.Topic,
					ev.TopicPartition.Partition,
					ev.TopicPartition.Offset)
			}
		case kafka.Error:
			log.Printf("Erro no produtor: %v", ev)
		}
	}
}

func sendMessage(producer *kafka.Producer, topic string, key string, message []byte) error {
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
		Key:            []byte(key),
	}, nil)

	if err != nil {
		return fmt.Errorf("Falha ao produzir: %w", err)
	}

	return nil
}
