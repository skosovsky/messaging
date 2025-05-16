package config

import (
	"os"
	"strings"
)

type AppConfig struct {
	KafkaBootstrapServers []string
	SchemaRegistryURL     string
}

func NewAppConfig() *AppConfig {
	bootstrapServers := getEnv("KAFKA_BOOTSTRAP_SERVERS", "127.0.0.1:9094,127.0.0.1:9095,127.0.0.1:9096")
	schemaRegistryURL := getEnv("SCHEMA_REGISTRY_URL", "http://schema-registry:8081")

	return &AppConfig{
		KafkaBootstrapServers: strings.Split(bootstrapServers, ","),
		SchemaRegistryURL:     schemaRegistryURL,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
