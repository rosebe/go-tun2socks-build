package v2ray

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	vcore "v2ray.com/core"
	vnet "v2ray.com/core/common/net"
	vsession "v2ray.com/core/common/session"

	"github.com/eycorsican/go-tun2socks/core"
	"github.com/xxf098/go-tun2socks-build/pool"
)

type tcpHandler struct {
	ctx context.Context
	v   *vcore.Instance
}

func (h *tcpHandler) relay(lhs net.Conn, rhs net.Conn) {
	go func() {
		buf := pool.NewBytes(pool.BufSize)
		io.CopyBuffer(rhs, lhs, buf)
		pool.FreeBytes(buf)
		lhs.Close()
		rhs.Close()
	}()
	buf := pool.NewBytes(pool.BufSize)
	io.CopyBuffer(lhs, rhs, buf)
	pool.FreeBytes(buf)
	lhs.Close()
	rhs.Close()
}

func NewTCPHandler(ctx context.Context, instance *vcore.Instance) core.TCPConnHandler {
	return &tcpHandler{
		ctx: ctx,
		v:   instance,
	}
}

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	dest := vnet.DestinationFromAddr(target)
	sid := vsession.NewID()
	ctx := vsession.ContextWithID(h.ctx, sid)
	c, err := vcore.Dial(ctx, h.v, dest)
	if err != nil {
		return errors.New(fmt.Sprintf("dial V proxy connection failed: %v", err))
	}
	go h.relay(conn, c)
	return nil
}
