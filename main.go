package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lovoo/goka"
	"github.com/riferrei/srclient"

	"messaging/blocker"
	"messaging/censor"
	"messaging/codec"
	"messaging/config"
	"messaging/emitter"
	"messaging/filter"
	"messaging/user"
)

func main() {
	appConfig := config.NewAppConfig()

	schemaRegistryClient := srclient.NewSchemaRegistryClient(appConfig.SchemaRegistryURL)
	messageCodec, err := codec.NewJSONSchema[user.Message](schemaRegistryClient, "messages-value", user.MessageSchema)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	blockerTopic := goka.Stream("blocked-users")
	blockerGroup := goka.Group("blocked-users")
	go blocker.Run(ctx, appConfig.KafkaBootstrapServers, blockerTopic, blockerGroup)

	denyWordsTopic := goka.Stream("deny-words")
	denyWordsGroup := goka.Group("deny-words")
	go censor.Run(ctx, appConfig.KafkaBootstrapServers, denyWordsTopic, denyWordsGroup)

	messagesTopic := goka.Stream("messages")
	filterTopic := goka.Stream("filtered-messages")
	filterGroup := goka.Group("filtered-messages")
	go filter.Run(ctx, appConfig.KafkaBootstrapServers, messagesTopic, messageCodec, filterTopic, filterGroup, blockerGroup, denyWordsGroup)

	go emitter.Run(ctx, appConfig.KafkaBootstrapServers, messagesTopic, messageCodec)

	<-signals
	log.Println("Interrupt signal received, shutting down...")
	cancel()
}
