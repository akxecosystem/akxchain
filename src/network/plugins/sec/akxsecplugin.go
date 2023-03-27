package sec

import (
	"github.com/libp2p/go-libp2p/core/sec"
	"go.dedis.ch/kyber/v3"
)

type AkxSecPlugin struct {
	t   sec.SecureTransport
	g   kyber.Group
	pub kyber.Point
}
