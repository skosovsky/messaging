package filter

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/lovoo/goka"
	"github.com/spaolacci/murmur3"

	"messaging/blocker"
	"messaging/censor"
	"messaging/codec"
	"messaging/user"
)

func Run(ctx context.Context, brokers []string, inputTopic goka.Stream, inputCodec goka.Codec, outputTopic goka.Stream, filterGroup, blockerGroup, denyWords goka.Group) {
	processorGraph := goka.DefineGroup(filterGroup,
		goka.Input(inputTopic, inputCodec, prepareProcess(outputTopic, blockerGroup, denyWords)),
		goka.Output(outputTopic, codec.JSON[user.Message]{}),
		goka.Join(goka.GroupTable(blockerGroup), codec.JSON[blocker.Values]{}),
		goka.Lookup(goka.GroupTable(denyWords), codec.JSON[censor.Value]{}),
	)

	processor, err := goka.NewProcessor(
		brokers,
		processorGraph,
		goka.WithHasher(murmur3.New32),
	)
	if err != nil {
		slog.ErrorContext(ctx, "new processor error", "error", err)

		return
	}

	err = processor.Run(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "run processor error", "error", err)
	}
}

func prepareProcess(outputTopic goka.Stream, groupBlocker, denyWords goka.Group) func(goka.Context, any) {
	return func(ctxProc goka.Context, msg any) {
		message, ok := msg.(*user.Message)
		if !ok || message == nil {

			return
		}

		// there is censorship here
		words := strings.Fields(message.Message)

		for i := range words {
			denyValueRaw := ctxProc.Lookup(goka.GroupTable(denyWords), words[i])
			if denyValueRaw != nil {
				var denyValue *censor.Value
				if denyValue, ok = denyValueRaw.(*censor.Value); ok && denyValue.Deny {
					words[i] = "***"
				}
			}
		}

		message.Message = strings.Join(words, " ")

		value := ctxProc.Join(goka.GroupTable(groupBlocker))
		if value == nil {
			ctxProc.Emit(outputTopic, ctxProc.Key(), message)

			return
		}

		blockValues, ok := value.(*blocker.Values)
		if !ok {

			return
		}

		// here is a ban check
		if slices.Contains(blockValues.BlockedIDs, message.UserID) {

			return
		}

		ctxProc.Emit(outputTopic, ctxProc.Key(), message)
	}
}
