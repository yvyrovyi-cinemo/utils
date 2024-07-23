package natslib

import (
	"context"
	"fmt"
	"net"
)

type dialer struct {
	ctx    context.Context
	dialer *net.Dialer
}

func newDialer(ctx context.Context) *dialer {
	return &dialer{
		ctx:    ctx,
		dialer: &net.Dialer{},
	}
}

func (d *dialer) Dial(network, address string) (net.Conn, error) {
	conn, err := d.dialer.DialContext(d.ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	return conn, nil
}
