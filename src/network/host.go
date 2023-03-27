package network

import (
	"bytes"
	"context"
	"errors"
	"github.com/ipfs/go-datastore"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	_ "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ma "github.com/multiformats/go-multiaddr"
	"sync"
)

type AKXNode struct {
	config    *NodeConfig
	Identity  peer.ID
	Host      host.Host
	PeerStore peerstore.Peerstore
}

type FakeDatastore struct {
	datastore.Batching
}

func NewNode(ctx context.Context, handler network.StreamHandler) (host.Host, *dht.IpfsDHT, error) {

	node := &AKXNode{}
	buf := bytes.NewBuffer(nil)
	priv, _, err := crypto.GenerateEd25519Key(buf)

	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/55456", "/ip6/::/tcp/55456"),
		libp2p.Identity(priv),
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
		libp2p.DefaultMuxers,
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.FallbackDefaults)
	if err != nil {
		panic(err)
	}
	node.Host = h
	node.Identity = h.ID()
	node.Host.SetStreamHandler(protocol.ID("/akxeco/1.0.0"), handler)
	dhtClient := dht.NewDHTClient(ctx, node.Host, &FakeDatastore{})
	// Define Bootstrap Nodes.
	peers := []string{
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	}

	// Convert Bootstap Nodes into usable addresses.
	BootstrapPeers := make(map[peer.ID]*peer.AddrInfo, len(peers))
	for _, addrStr := range peers {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			return node.Host, dhtClient, err
		}
		pii, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return node.Host, dhtClient, err
		}
		pi, ok := BootstrapPeers[pii.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: pii.ID}
			BootstrapPeers[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, pii.Addrs...)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	lock := sync.Mutex{}
	count := 0
	wg.Add(len(BootstrapPeers))
	for _, peerInfo := range BootstrapPeers {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			err := node.Host.Connect(ctx, *peerInfo)
			if err == nil {
				lock.Lock()
				count++
				lock.Unlock()

			}
		}(peerInfo)
	}
	wg.Wait()

	if count < 1 {
		return node.Host, dhtClient, errors.New("unable to bootstrap libp2p node")
	}

	return node.Host, dhtClient, nil

}
