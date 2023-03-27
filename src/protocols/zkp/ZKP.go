package zkp

import (
	"github.com/cloudflare/circl/zk/dleq"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type ProtocolZKP struct {
	ID       protocol.ID
	prover   dleq.Prover
	verifier dleq.Verifier
}
