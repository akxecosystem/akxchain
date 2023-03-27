package network

import (
	"github.com/akxecosystem/akxchain/libs/kem"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"sync"
)

type HandshakeProtocol struct {
	inputCh       chan []byte
	outputCh      chan []byte
	dataLock      sync.Mutex
	ephemeralKeys *kem.Keys
	challenge     chan []byte   // 32 byte challenge to be sent here
	done          chan struct{} // call when done and succeeded
	closeConn     chan struct{} // call when failed just close the connection
}

const HS_PROTOCOL_ID = protocol.ID("k_handshake")

func (khs *HandshakeProtocol) Init(peer1, peer2 peer.ID) {

}
