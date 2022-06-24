package main

import (
	"context"
	"io"
	mrand "math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"

	"github.com/libp2p/go-libp2p"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

func main() {
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
	ha, err := makeBasicHost(123456789)
	if err != nil {
		panic(err)
	}
	defer ha.Close()

	ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
		defer s.Close()
		s.SetDeadline(time.Now().Add(time.Second))
		_, err := io.Copy(s, s)
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}
	})

	_, err = os.Stdout.WriteString("running")
	if err != nil {
		panic(err)
	}
	err = os.Stdout.Close()
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
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
		libp2p.ListenAddrStrings("/ip4/127.13.37.42/tcp/12345", "/ip4/127.13.37.42/udp/12345/quic"),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(quic.NewTransport),
	}

	return libp2p.New(opts...)
}
