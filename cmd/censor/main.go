package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"messaging/censor"
	"messaging/codec"

	"github.com/lovoo/goka"
	"github.com/spaolacci/murmur3"
)

// go run cmd/censor/main.go -word text -action deny -stream deny-words
// go run cmd/censor/main.go -word text -action permit -stream deny-words

func main() {
	ctx := context.Background()

	var (
		word   = flag.String("word", "", "word for action")
		action = flag.String("action", "", "permit or deny only")
		broker = flag.String("broker", "localhost:9094", "bootstrap broker server")
		stream = flag.String("stream", "", "stream name")
	)
	flag.Parse()

	if *word == "" || *action == "" || *stream == "" {
		slog.ErrorContext(ctx, "-word, -action, -stream must be set")
		os.Exit(1)
	}

	var actionCode censor.Action
	switch *action {
	case "permit":
		actionCode = censor.ShouldPermit
	case "deny":
		actionCode = censor.ShouldDeny
	default:
		slog.ErrorContext(ctx, "-action must be set only permit or deny")
		os.Exit(1)
	}

	runEmitter(ctx, []string{*broker}, goka.Stream(*stream), *word, actionCode)
}

func runEmitter(ctx context.Context, brokers []string, topic goka.Stream, word string, action censor.Action) {
	emitter, err := goka.NewEmitter(
		brokers,
		topic,
		codec.JSON[censor.Event]{},
		goka.WithEmitterHasher(murmur3.New32),
	)
	if err != nil {
		slog.ErrorContext(ctx, "new emitter error", "error", err)
	}
	defer emitter.Finish()

	err = emitter.EmitSync(word, &censor.Event{
		Word:   word,
		Action: action,
	})
	if err != nil {
		slog.ErrorContext(ctx, "emit sync error", "error", err)
	}
}
