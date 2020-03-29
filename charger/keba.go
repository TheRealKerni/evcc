package charger

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/andig/evcc/api"
)

const (
	udpBufferSize = 1024
	updTimeout    = time.Second
)

// Keba is an api.Charger implementation with configurable getters and setters.
type Keba struct {
	log  *api.Logger
	conn string
}

// NewKebaFromConfig creates a new configurable charger
func NewKebaFromConfig(log *api.Logger, other map[string]interface{}) api.Charger {
	cc := struct{ URI string }{}
	api.DecodeOther(log, other, &cc)

	return NewKeba(cc.URI)
}

// NewKeba creates a new charger
func NewKeba(conn string) api.Charger {
	log := api.NewLogger("keba")

	return &Keba{
		log:  log,
		conn: conn,
	}
}

// client wraps the whole functionality of a UDP client that sends
// a message and waits for a response coming back from the server
// that it initially targetted.
// https://ops.tips/blog/udp-client-and-server-in-go/#sending-udp-packets-using-go
func (m *Keba) client(ctx context.Context, address string, reader io.Reader) error {
	// Resolve the UDP address so that we can make use of DialUDP
	// with an actual IP and port instead of a name (in case a
	// hostname is specified).
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	// Although we're not in a connection-oriented transport,
	// the act of `dialing` is analogous to the act of performing
	// a `connect(2)` syscall for a socket of type SOCK_DGRAM:
	// - it forces the underlying socket to only read and write
	//   to and from a specific remote address.
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}

	// Closes the underlying file descriptor associated with the
	// socket so that it no longer refers to any file.
	defer conn.Close()

	doneChan := make(chan error, 1)

	go func() {
		// It is possible that this action blocks, although this
		// should only occur in very resource-intensive situations:
		// - when you've filled up the socket buffer and the OS
		//   can't dequeue the queue fast enough.
		n, err := io.Copy(conn, reader)
		if err != nil {
			doneChan <- err
			return
		}

		m.log.TRACE.Printf("packet written: bytes=%d\n", n)

		buffer := make([]byte, udpBufferSize)

		// Set a deadline for the ReadOperation so that we don't
		// wait forever for a server that might not respond on
		// a reasonable amount of time.
		deadline := time.Now().Add(updTimeout)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			doneChan <- err
			return
		}

		nRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			doneChan <- err
			return
		}

		m.log.TRACE.Printf("packet received: bytes=%d from=%s\n", nRead, addr.String())

		doneChan <- nil
	}()

	select {
	case <-ctx.Done():
		m.log.TRACE.Println("cancelled")
		err = ctx.Err()
	case err = <-doneChan:
	}

	return err
}

// Status implements the Charger.Status interface
func (m *Keba) Status() (api.ChargeStatus, error) {
	return api.ChargeStatus("A"), errors.New("not implemented")
}

// Enabled implements the Charger.Enabled interface
func (m *Keba) Enabled() (bool, error) {
	return false, errors.New("not implemented")
}

// Enable implements the Charger.Enable interface
func (m *Keba) Enable(enable bool) error {
	return errors.New("not implemented")
}

// MaxCurrent implements the Charger.MaxCurrent interface
func (m *Keba) MaxCurrent(current int64) error {
	return errors.New("not implemented")
}
