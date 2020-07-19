package provider

import (
	"context"
	"io"
	"net"
	"sync"
	"time"
)

type StartStopConfig struct {
	numConn int

	SourceAddr net.Addr
	TargetAddr net.Addr
	KeepAlive  time.Duration

	active    bool
	idleTimer *time.Timer
	IdleDur   time.Duration

	Start func() error
	Stop  func() error
}

func (ssc *StartStopConfig) listen(l net.Listener, connCh chan<- net.Conn, errCh chan<- error) {
	for {
		c, err := l.Accept()
		if err != nil {
			errCh <- err
			return
		}
		connCh <- c
	}
}

func (ssc *StartStopConfig) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a new (inactive) timer
	ssc.idleTimer = time.NewTimer(ssc.IdleDur)
	if !ssc.idleTimer.Stop() {
		<-ssc.idleTimer.C
	}

	// Listen for incoming connections
	errCh := make(chan error)
	connCh := make(chan net.Conn)
	closeCh := make(chan struct{})
	lc := net.ListenConfig{
		KeepAlive: ssc.KeepAlive,
	}
	l, err := lc.Listen(ctx, ssc.SourceAddr.Network(), ssc.SourceAddr.String())
	if err != nil {
		return err
	}
	defer l.Close()
	go ssc.listen(l, connCh, errCh)

	// Handle connections and start/stop functionality
	for {
		select {
		case c := <-connCh:
			// Clear the idle timer
			if !ssc.idleTimer.Stop() {
				select {
				case <-ssc.idleTimer.C:
				default:
					break
				}
			}

			// Ensure the server is active
			if !ssc.active {
				if err := ssc.Start(); err != nil {
					return err
				}
				ssc.active = true
			}

			ssc.numConn++
			go func() {
				ssc.serve(ctx, c)
				closeCh <- struct{}{}
			}()
		case <-closeCh:
			// Stop and reset the timer
			if !ssc.idleTimer.Stop() {
				select {
				case <-ssc.idleTimer.C:
				default:
					break
				}
			}
			ssc.idleTimer.Reset(ssc.IdleDur)

			ssc.numConn--
		case <-ssc.idleTimer.C:
			// Stop the server
			if err := ssc.Stop(); err != nil {
				return err
			}
			ssc.active = false
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (ssc *StartStopConfig) serve(ctx context.Context, sc net.Conn) {
	var d net.Dialer
	tc, err := d.DialContext(ctx, ssc.TargetAddr.Network(), ssc.TargetAddr.String())
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(tc, sc)
		sc.Close()
		tc.Close()
		wg.Done()
	}()
	go func() {
		io.Copy(sc, tc)
		tc.Close()
		sc.Close()
		wg.Done()
	}()
	wg.Wait()
}
