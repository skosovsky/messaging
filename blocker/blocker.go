package blocker

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/lovoo/goka"
	"github.com/spaolacci/murmur3"

	"messaging/codec"
)

type Action int

const (
	ShouldUnblock Action = iota + 1
	ShouldBlock
)

type Event struct {
	AuthorID    int
	RecipientID int
	Action      Action
}

type Values struct {
	BlockedIDs []int
}

func Run(ctx context.Context, brokers []string, topic goka.Stream, group goka.Group) {
	processorGraph := goka.DefineGroup(group,
		goka.Input(topic, codec.JSON[Event]{}, process),
		goka.Persist(codec.JSON[Values]{}),
	)

	processor, err := goka.NewProcessor(
		brokers,
		processorGraph,
		goka.WithHasher(murmur3.New32),
	)
	if err != nil {
		slog.ErrorContext(ctx, "new processor error", "error", err)
	}

	err = processor.Run(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "run processor error", "error", err)
	}
}

func process(ctxProc goka.Context, msg any) {
	val := ctxProc.Value()

	stored, ok := val.(*Values)
	if !ok || stored == nil {
		stored = &Values{}
	}

	event, ok := msg.(*Event)
	if !ok || event == nil || event.AuthorID == event.RecipientID {
		return
	}

	switch event.Action {
	case ShouldUnblock:
		if !slices.Contains(stored.BlockedIDs, event.RecipientID) {
			return
		}

		stored.BlockedIDs = slices.DeleteFunc(stored.BlockedIDs, func(id int) bool {
			return id == event.RecipientID
		})

		ctxProc.SetValue(stored)
		fmt.Printf("[proc] user: %s, unblocked %d\n", ctxProc.Key(), event.RecipientID) //nolint:forbidigo // example

	case ShouldBlock:
		if slices.Contains(stored.BlockedIDs, event.RecipientID) {
			return
		}

		stored.BlockedIDs = append(stored.BlockedIDs, event.RecipientID)

		ctxProc.SetValue(stored)
		fmt.Printf("[proc] user: %s, blocked %d\n", ctxProc.Key(), event.RecipientID) //nolint:forbidigo // example
	}
}
