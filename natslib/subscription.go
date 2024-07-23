package natslib

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Subscription struct {
	natsSubscription *nats.Subscription
}

func (s *Subscription) Receive(ctx context.Context) ([]byte, error) {
	msg, err := s.natsSubscription.NextMsgWithContext(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err //nolint:wrapcheck // intentionally pass as-is
		}

		// nats.ErrBadSubscription is used by the nats library when the subscription is closed or is nil.
		// We assume "closed" in both cases.
		if errors.Is(err, nats.ErrConnectionClosed) || errors.Is(err, nats.ErrBadSubscription) {
			return nil, ErrClosed
		}

		return nil, fmt.Errorf("subscription next msg: %w", err)
	}
	return msg.Data, nil
}
