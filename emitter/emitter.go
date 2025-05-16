package emitter

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/lovoo/goka"
	"github.com/spaolacci/murmur3"

	"messaging/user"
)

func Run(ctx context.Context, brokers []string, topic goka.Stream, codec goka.Codec) {
	const emitInterval = 3 * time.Second

	emitter, err := goka.NewEmitter(brokers, topic, codec,
		goka.WithEmitterHasher(murmur3.New32),
	)
	if err != nil {
		slog.ErrorContext(ctx, "emitter error", "error", err)
	}
	defer emitter.Finish()

	ticker := time.NewTicker(emitInterval)
	defer ticker.Stop()

	for range ticker.C {
		senderID := rand.IntN(3) + 1   //nolint:gosec,mnd // example
		receiverID := rand.IntN(3) + 1 //nolint:gosec,mnd // example

		var text string
		if rand.IntN(10)%2 == 0 { //nolint:gosec,mnd // example
			text = "some text"
		} else {
			text = "another test words"
		}

		fakeMessage := &user.Message{
			UserID:      senderID,
			RecipientID: receiverID,
			Message:     fmt.Sprintf("%s %d", text, rand.IntN(10)), //nolint:gosec,mnd // example
			CreatedAt:   time.Now(),
		}

		err = emitter.EmitSync(strconv.FormatInt(int64(receiverID), 10), fakeMessage)
		if err != nil {
			slog.ErrorContext(ctx, "emitter error", "error", err)
		}
	}
}
