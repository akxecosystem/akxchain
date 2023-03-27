package sec

import (
	"bufio"
	"context"
	ecies "github.com/ecies/go/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/sec/insecure/pb"
	"github.com/libp2p/go-msgio"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/util/random"
	"go.dedis.ch/kyber/v3/xof/blake2xb"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

type secureSession struct {
	suite       *edwards25519.SuiteEd25519
	initiator   bool
	checkPeerID bool
	localID     peer.ID
	localKey    crypto.PrivKey
	remoteID    peer.ID
	remoteKey   crypto.PubKey

	readLock  sync.Mutex
	writeLock sync.Mutex

	insecureConn   net.Conn
	insecureReader *bufio.Reader // to cushion io read syscalls
	// we don't buffer writes to avoid introducing latency; optimisation possible. // TODO revisit

	qseek                                                int     // queued bytes seek value.
	qbuf                                                 []byte  // queued bytes buffer.
	rlen                                                 [2]byte // work buffer to read in the incoming message length.
	initiatorEarlyDataHandler, responderEarlyDataHandler EarlyDataHandler
	sharedSecret                                         []byte
	pub                                                  []byte
	priv                                                 []byte
	// ConnectionState holds state information releated to the secureSession entity.
	connectionState network.ConnectionState
}

type Transport struct{}

type EarlyDataHandler struct{}

func NewSecureSession(tpt *Transport, ctx context.Context, insecure net.Conn, remote peer.ID, initiatorEDH, responderEDH EarlyDataHandler, initiator, checkPeerID bool) (*secureSession, error) {
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		panic(err)
	}
	s := &secureSession{}
	s.localKey = priv
	s.remoteKey = pub
	s.insecureConn = insecure
	s.insecureReader = bufio.NewReader(insecure)
	s.initiator = initiator
	s.remoteID = remote
	s.initiatorEarlyDataHandler = initiatorEDH
	s.responderEarlyDataHandler = responderEDH
	s.checkPeerID = checkPeerID
	s.NewKyberKeyPair()

	respCh := make(chan error, 1)
	go func() {
		respCh <- s.runHandshake(ctx)
	}()

	select {
	case err := <-respCh:
		if err != nil {
			_ = s.insecureConn.Close()
		}
		return s, err

	case <-ctx.Done():
		// If the context has been cancelled, we close the underlying connection.
		// We then wait for the handshake to return because of the first error it encounters
		// so we don't return without cleaning up the go-routine.
		_ = s.insecureConn.Close()
		<-respCh
		return nil, ctx.Err()
	}

	return s, nil

}

func (s *secureSession) NewKyberKeyPair() {
	rng := blake2xb.New(nil)
	suite := edwards25519.NewBlakeSHA256Ed25519()
	s.suite = suite
	privKey := s.suite.Scalar().Pick(rng)
	pubKey := s.suite.Point().Mul(privKey, nil)
	s.pub, _ = pubKey.MarshalBinary()
	s.priv, _ = privKey.MarshalBinary()

}

func (s *secureSession) GenerateSharedSecret() {

}

func (s *secureSession) runHandshake(ctx context.Context) error {

	pub := s.suite.Point()
	var pubKey []byte
	pub.UnmarshalBinary(pubKey)

	encKey, _ := ecies.GenerateKey()
	msg := s.generateNewNonce()
	_, err := ecies.Encrypt(encKey.PublicKey, msg)
	if err != nil {
		panic(err)
	}
	localPub, _ := s.localID.ExtractPublicKey()
	keyMsg, err := crypto.PublicKeyToProto(localPub)

	msgProto := &pb.Exchange{
		Id:     []byte(s.LocalPeer()),
		Pubkey: keyMsg,
	}
	_, err = readWriteMsg(s.insecureConn, msgProto)
	return nil
}

func (s *secureSession) generateNewNonce() []byte {
	localNonceRng := blake2xb.New(nil)
	var nonce []byte
	localNonceRng.Read(nonce)
	return nonce
}

func (s *secureSession) Error() string {
	return ""
}

func (s *secureSession) LocalAddr() net.Addr {
	return s.insecureConn.LocalAddr()
}

func (s *secureSession) LocalPeer() peer.ID {
	return s.localID
}

func (s *secureSession) LocalPublicKey() crypto.PubKey {
	return s.localKey.GetPublic()
}

func (s *secureSession) RemoteAddr() net.Addr {
	return s.insecureConn.RemoteAddr()
}

func (s *secureSession) RemotePeer() peer.ID {
	return s.remoteID
}

func (s *secureSession) RemotePublicKey() crypto.PubKey {
	return s.remoteKey
}

func (s *secureSession) ConnState() network.ConnectionState {
	return s.connectionState
}

func (s *secureSession) SetDeadline(t time.Time) error {
	return s.insecureConn.SetDeadline(t)
}

func (s *secureSession) SetReadDeadline(t time.Time) error {
	return s.insecureConn.SetReadDeadline(t)
}

func (s *secureSession) SetWriteDeadline(t time.Time) error {
	return s.insecureConn.SetWriteDeadline(t)
}

func (s *secureSession) Close() error {
	return s.insecureConn.Close()
}

type Handshake struct {
	group   kyber.Group
	message []byte
}

func (h *Handshake) initialize(pubKey kyber.Point, message []byte) (K, C kyber.Point, remainder []byte) {
	// Embed the message (or as much of it as will fit) into a curve point.
	M := h.group.Point().Embed(message, random.New())
	max := h.group.Point().EmbedLen()
	if max > len(message) {
		max = len(message)
	}
	remainder = message[max:]
	// encrypt the point to produce ciphertext (K,C).
	k := h.group.Scalar().Pick(random.New()) // ephemeral private key
	K = h.group.Point().Mul(k, nil)          // ephemeral DH public key
	S := h.group.Point().Mul(k, pubKey)      // ephemeral DH shared secret
	C = S.Add(S, M)                          // message blinded with secret
	return
}

// read and write a message at the same time.
func readWriteMsg(rw io.ReadWriter, out *pb.Exchange) (*pb.Exchange, error) {
	const maxMessageSize = 1 << 16

	outBytes, err := proto.Marshal(out)
	if err != nil {
		return nil, err
	}
	wresult := make(chan error)
	go func() {
		w := msgio.NewVarintWriter(rw)
		wresult <- w.WriteMsg(outBytes)
	}()

	r := msgio.NewVarintReaderSize(rw, maxMessageSize)
	b, err1 := r.ReadMsg()

	// Always wait for the read to finish.
	err2 := <-wresult

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		r.ReleaseMsg(b)
		return nil, err2
	}
	inMsg := new(pb.Exchange)
	err = proto.Unmarshal(b, inMsg)
	return inMsg, err
}
