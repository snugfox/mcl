package provider

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type StartStopConfig struct {
	SourceAddr net.Addr
	TargetAddr net.Addr
	KeepAlive  time.Duration

	IdleDur time.Duration

	RunFunc  func(ctx context.Context) error
	StopFunc func(ctx context.Context) error
}

func disableTimer(t *time.Timer, waitCh bool) {
	if !t.Stop() {
		if waitCh {
			<-t.C
		} else {
			select {
			case <-t.C:
			default:
				break
			}
		}
	}
}

func (ssc *StartStopConfig) listen(l net.Listener, connCh chan<- net.Conn) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		connCh <- c
	}
}

func (ssc *StartStopConfig) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a new (inactive) timer
	idleTimer := time.NewTimer(ssc.IdleDur)
	disableTimer(idleTimer, true)
	defer disableTimer(idleTimer, false)

	// Listen for incoming connections
	lErrCh := make(chan error)    // Listener errors (implies termination)
	connCh := make(chan net.Conn) // New connections
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, ssc.SourceAddr.Network(), ssc.SourceAddr.String())
	if err != nil {
		return err
	}
	defer l.Close()
	go func() {
		lErrCh <- ssc.listen(l, connCh)
	}()

	// Handle incoming connections and start/stop functionality
	var sActive bool
	var nConn int
	cErrCh := make(chan error) // Per connection errors (can be nil; implies termination)
	sErrCh := make(chan error) // Server error (can be nil; implies termination)
	for {
		select {
		case c := <-connCh:
			log.Println("conn")
			// Disable the idle timer
			disableTimer(idleTimer, false)

			// Start the server if it isn't already
			if !sActive {
				go func() {
					sErrCh <- ssc.RunFunc(ctx)
				}()
				sActive = true
			}

			// Serve the connection
			nConn++
			go func() {
				cErrCh <- ssc.serve(ctx, c)
			}()
		case <-idleTimer.C:
			log.Println("timer")
			// The timer should only trigger when the server is active
			if !sActive {
				panic("idle timer expired with an inactive server")
			}

			// Stop the server
			log.Println("stop func")
			if err := ssc.StopFunc(ctx); err != nil {
				return err
			}
			<-sErrCh // Wait for ssc.RunFunc to return
			sActive = false
		case err := <-lErrCh:
			log.Println("listener term")
			return err
		case err := <-sErrCh:
			log.Println("server term")
			if err != nil {
				return err
			}
			sActive = false
			disableTimer(idleTimer, false)
		case <-cErrCh: // TODO: Handle error
			log.Println("conn term")
			nConn--
			if nConn == 0 {
				// Stop and reset the timer
				disableTimer(idleTimer, false)
				idleTimer.Reset(ssc.IdleDur)
			}
		case <-ctx.Done():
			log.Println("ctx term")
			return ctx.Err()
		}
	}
}

func (ssc *StartStopConfig) serve(ctx context.Context, sc net.Conn) error {
	var d net.Dialer
	var tc net.Conn
	for tc == nil {
		tc, _ = d.DialContext(ctx, ssc.TargetAddr.Network(), ssc.TargetAddr.String())
		if err := ctx.Err(); err != nil {
			return err
		}
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

	// TODO: Handle ctx
	return nil
}
