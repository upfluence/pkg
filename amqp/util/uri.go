package util

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/streadway/amqp"

	dpeer "github.com/upfluence/pkg/discovery/peer"
	lpeer "github.com/upfluence/pkg/peer"
)

func peerTable(p *lpeer.Peer) amqp.Table {
	if p == nil {
		return amqp.Table{}
	}

	return amqp.Table{
		"upfluence-unit-name":    p.InstanceName,
		"upfluence-app-name":     p.AppName,
		"upfluence-project-name": p.ProjectName,
		"upfluence-env":          p.Environment,
		"upfluence-version":      lpeer.SerializeVersion(p.Version),
	}
}

func peerURI(p *dpeer.Peer) string {
	if strings.HasPrefix("amqp://", p.Addr) {
		return p.Addr
	}

	return fmt.Sprintf("amqp://guest:guest@%s/%%2f", p.Addr)
}

func Dial(ctx context.Context, p *dpeer.Peer, l *lpeer.Peer, name string) (*amqp.Connection, error) {
	var (
		d net.Dialer

		table = peerTable(l)
	)

	if name != "" {
		table["upfluence-connection-name"] = name
	}

	return amqp.DialConfig(
		peerURI(p),
		amqp.Config{
			Dial: func(network, addr string) (net.Conn, error) {
				return d.DialContext(ctx, network, addr)
			},
			Properties: table,
		},
	)
}
