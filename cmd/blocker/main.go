package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strconv"

	"github.com/lovoo/goka"
	"github.com/spaolacci/murmur3"

	"messaging/blocker"
	"messaging/codec"
)

// go run cmd/blocker/main.go -author 1 -recipient 2 -action block -stream blocked-users
// go run cmd/blocker/main.go -author 1 -recipient 2 -action unblock -stream blocked-users

func main() {
	ctx := context.Background()

	var (
		author    = flag.Int("author", 0, "author who is blocking")
		recipient = flag.Int("recipient", 0, "recipient who is blocked")
		action    = flag.String("action", "", "unblock or block only")
		broker    = flag.String("broker", "localhost:9094", "bootstrap broker server")
		stream    = flag.String("stream", "", "stream name")
	)
	flag.Parse()

	if *author == 0 || *recipient == 0 || *action == "" || *stream == "" {
		slog.ErrorContext(ctx, "-author, -recipient, -action, -stream must be set")
		os.Exit(1)
	}

	if *author == *recipient {
		slog.ErrorContext(ctx, "-author and -recipient must be different")
		os.Exit(1)
	}

	var actionCode blocker.Action
	switch *action {
	case "unblock":
		actionCode = blocker.ShouldUnblock
	case "block":
		actionCode = blocker.ShouldBlock
	default:
		slog.ErrorContext(ctx, "-action must be set only unblock or block")
		os.Exit(1)
	}

	runEmitter(ctx, []string{*broker}, goka.Stream(*stream), *author, *recipient, actionCode)
}

func runEmitter(ctx context.Context, brokers []string, topic goka.Stream, authorID, recipientID int, action blocker.Action) {
	emitter, err := goka.NewEmitter(
		brokers,
		topic,
		codec.JSON[blocker.Event]{},
		goka.WithEmitterHasher(murmur3.New32),
	)
	if err != nil {
		slog.ErrorContext(ctx, "new emitter error", "error", err)
	}
	defer emitter.Finish()

	key := strconv.FormatInt(int64(authorID), 10)
	err = emitter.EmitSync(key, &blocker.Event{
		AuthorID:    authorID,
		RecipientID: recipientID,
		Action:      action,
	})
	if err != nil {
		slog.ErrorContext(ctx, "emit sync error", "error", err)
	}
}
