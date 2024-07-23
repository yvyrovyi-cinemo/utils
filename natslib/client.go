package natslib

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

type (
	Client struct {
		conn           *nats.Conn
		logger         *slog.Logger
		subCheckPeriod time.Duration
	}

	Config struct {
		URL            string        `env:"URL" json:"url" yaml:"url"`
		DrainTimeout   time.Duration `env:"DRAIN_TIMEOUT" envDefault:"10s" json:"drain_timeout" yaml:"drain_timeout" default:"10s"`
		SubCheckPeriod time.Duration `env:"SUB_CHECK_PERIOD" envDefault:"10s" json:"sub_check_period" yaml:"sub_check_period" default:"10s"`
		MaxReconnect   int           `env:"MAX_RECONNECT" envDefault:"60" json:"max_reconnect" yaml:"max_reconnect" default:"60"`
		ReconnectWait  time.Duration `env:"RECONNECT_WAIT" envDefault:"1s" json:"reconnect_wait" yaml:"reconnect_wait" default:"1s"`

		// SubscriptionPendingMessagesLimit is the maximum number of pending messages for a subscription.
		SubscriptionPendingMessagesLimit int `env:"SUBSCRIPTION_PENDING_MESSAGES_LIMIT" envDefault:"1000" json:"subscription_pending_messages_limit" yaml:"subscription_pending_messages_limit" default:"1000"`
	}
)

var (
	// ErrPublishFailed is an error returned when pubsub fails to publish message
	ErrPublishFailed = errors.New("publish failed")

	// ErrClosed is an error returned when pubsub instance is closed.
	ErrClosed = errors.New("pubsub is closed")
)

func Connect(ctx context.Context, config Config, logger *slog.Logger) (*Client, error) {
	natsOptions := []nats.Option{
		nats.SetCustomDialer(newDialer(ctx)),
		nats.DrainTimeout(config.DrainTimeout),

		// Set the number of reconnect attempts and the time to wait
		// between reconnect attempts to the NATS server.
		nats.MaxReconnects(config.MaxReconnect),
		nats.ReconnectWait(config.ReconnectWait),

		// set the maximum number of pending messages for a subscription.
		nats.SyncQueueLen(config.SubscriptionPendingMessagesLimit),

		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			logger.Error("nats package error", "error", err)
		}),

		// Reports the error in case when the connection is disconnected.
		// In that case subscription.IsValid() returns true until max reconnect
		// and reconnect wait will be exceeded, and we have to report the error
		// immediately in that handler.
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if errors.Is(err, io.EOF) {
				// a client losing the connection is not an unexpected error
				return
			}
			logger.Error("nats package disconnect error", "error", err)
		}),
	}

	nc, err := nats.Connect(config.URL, natsOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	return &Client{
		conn:           nc,
		logger:         logger,
		subCheckPeriod: config.SubCheckPeriod,
	}, nil
}

// Close closes the NATS connection.
func (c *Client) Close() {
	c.conn.Close()
}

// Publish publishes a message to the NATS server.
func (c *Client) Publish(ctx context.Context, topic string, payload []byte) error {
	// Check if the connection still has connection to the NATS server.
	// In that case connection is not connected to the NATS server but is not closed
	// and tries to reconnect until reaching the max reconnect attempts.
	// If not connected, then all messages will be saved in the buffer(default size 8 MB) and will be
	// sent to the NATS server when the connection will be established.
	// We don't want to return an error in that case, but we have to log it.
	if !c.conn.IsConnected() {
		return errors.New("nats is not connected")
	}

	if err := c.conn.Publish(topic, payload); err != nil {
		if errors.Is(err, nats.ErrConnectionClosed) {
			return ErrClosed
		}

		return fmt.Errorf("%w: %v", ErrPublishFailed, err)
	}

	return nil
}

// Subscribe subscribes to the NATS server.
func (c *Client) Subscribe(ctx context.Context, topic string) (*Subscription, error) {
	natsSubscription, err := c.conn.SubscribeSync(topic)
	if err != nil {
		if errors.Is(err, nats.ErrConnectionClosed) {
			return nil, ErrClosed
		}
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	sub := Subscription{natsSubscription: natsSubscription}

	go c.waitSubscriptionCancel(ctx, natsSubscription)

	return &sub, nil
}

func (c *Client) waitSubscriptionCancel(
	ctx context.Context,
	subscription *nats.Subscription,
) {
	ticker := time.NewTicker(c.subCheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			err := subscription.Unsubscribe()
			if err != nil &&
				!errors.Is(err, nats.ErrConnectionClosed) &&
				!errors.Is(err, nats.ErrConnectionDraining) {

				c.logger.Error("failed to unsubscribe", "error", err)
			}
			return
		case <-ticker.C:
			if !subscription.IsValid() {
				return
			}
		}
	}
}
