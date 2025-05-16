package censor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lovoo/goka"
	"github.com/spaolacci/murmur3"

	"messaging/codec"
)

type Action int

const (
	ShouldPermit Action = iota + 1
	ShouldDeny
)

type Event struct {
	Word   string
	Action Action
}

type Value struct {
	Deny bool
}

func Run(ctx context.Context, brokers []string, topic goka.Stream, group goka.Group) {
	processorGraph := goka.DefineGroup(group,
		goka.Input(topic, codec.JSON[Event]{}, process),
		goka.Persist(codec.JSON[Value]{}),
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

	stored, ok := val.(*Value)
	if !ok || stored == nil {
		stored = &Value{
			Deny: false,
		}
	}

	event, ok := msg.(*Event)
	if !ok || event == nil {
		return
	}

	switch event.Action {
	case ShouldPermit:
		if !stored.Deny {
			return
		}

		stored.Deny = false

		ctxProc.SetValue(stored)
		fmt.Printf("[proc] word: %s, permit\n", ctxProc.Key()) //nolint:forbidigo // example

	case ShouldDeny:
		if stored.Deny {
			return
		}

		stored.Deny = true

		ctxProc.SetValue(stored)
		fmt.Printf("[proc] word: %s, deny\n", ctxProc.Key()) //nolint:forbidigo // example
	}
}
