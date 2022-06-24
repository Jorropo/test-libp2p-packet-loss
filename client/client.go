package main

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"io"
	mrand "math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"

	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

const testSize = 1024 * 1024 * 16 // 16MiB

func main() {
	pi, err := peer.AddrInfoFromString(os.Args[1])
	if err != nil {
		panic(err)
	}

	data := make([]byte, testSize)
	_, err = io.ReadFull(crand.Reader, data)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	// Make a host that listens on the given multiaddress
	ha, err := makeBasicHost(987654321)
	if err != nil {
		panic(err)
	}
	defer ha.Close()

	err = ha.Connect(ctx, *pi)
	if err != nil {
		panic(err)
	}

	s, err := ha.NewStream(ctx, pi.ID, "/echo/1.0.0")
	if err != nil {
		panic(err)
	}
	defer s.Close()
	s.SetDeadline(time.Now().Add(time.Second))
	var wait sync.Mutex
	wait.Lock()
	var red []byte
	go func() {
		defer wait.Unlock()
		var err error
		red, err = io.ReadAll(s)
		if err != nil {
			panic(err)
		}
	}()
	_, err = s.Write(data)
	if err != nil {
		panic(err)
	}
	err = s.CloseWrite()
	if err != nil {
		panic(err)
	}

	wait.Lock()

	if !bytes.Equal(data, red) {
		panic("data received != data sent")
	}
}

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It won't encrypt the connection if insecure is true.
func makeBasicHost(randseed int64) (host.Host, error) {
	r := mrand.New(mrand.NewSource(randseed))

	// Generate a key pair for this host. We will use it at least
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(quic.NewTransport),
	}

	return libp2p.New(opts...)
}
