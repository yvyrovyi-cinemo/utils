package restartclient

import (
	"context"
	"log/slog"
	"time"

	"git.internal.cinemo.com/~yvyrovyi/utils/duration"
)

type (
	Client struct {
		Config
		cancelFunc        func()
		coordinatorClient CoordinatorClient
		logger            *slog.Logger
	}

	Config struct {
		Enable          bool                    `env:"ENABLE" json:"enable"yaml:"enable"`
		GroupName       string                  `env:"GROUP_NAME" json:"group_name" yaml:"group_name"`
		RequestPeriod   duration.ConfigDuration `env:"REQUEST_PERIOD" json:"request_period" yaml:"request_period"`
		CoordinatorHost string                  `env:"COORDINATOR_HOST" json:"coordinator_host" yaml:"coordinator_host"`
		DelayOnStart    duration.ConfigDuration `env:"DELAY_ON_START" json:"delay_on_start" yaml:"delay_on_start"`
	}
)

const (
	DefaultRequestPeriod = 30 * time.Second
	DefaultDelayOnStart  = 15 * time.Second
)

type (
	CoordinatorClient interface {
		MustTerminate(ctx context.Context, groupName string) bool
	}
)

func New(config Config, cancelFunc func(), coordinatorClient CoordinatorClient, logger *slog.Logger) *Client {

	if config.RequestPeriod.Duration == 0 {
		config.RequestPeriod = duration.ConfigDuration{Duration: DefaultRequestPeriod}
	}

	if config.DelayOnStart.Duration == 0 {
		config.DelayOnStart = duration.ConfigDuration{Duration: DefaultDelayOnStart}
	}

	return &Client{
		Config:            config,
		cancelFunc:        cancelFunc,
		coordinatorClient: coordinatorClient,
		logger:            logger.With("component", "restart-client"),
	}
}

func (c *Client) Run(ctx context.Context) error {
	if !c.Enable {
		<-ctx.Done()
		return nil
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	case <-time.After(c.DelayOnStart.Duration):
	}

	for {
		if ok := c.coordinatorClient.MustTerminate(ctx, c.GroupName); ok {
			c.logger.Info("coordinator sent term signal")
			c.cancelFunc()
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		case <-time.After(c.DelayOnStart.Duration):
		}
	}
}
