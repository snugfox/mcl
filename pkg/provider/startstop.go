package provider

import (
	"context"
	"io"
	"log"
	"net"
	"time"
)

// An ErrStartStopMissingField indicates that a required start-stop parameter
// is unset.
type ErrStartStopMissingField string

func (e ErrStartStopMissingField) Error() string {
	return "required start-stop field " + string(e) + "is unset"
}

// A StartStopConfig defines parameters for running a server or other process
// that automatically starts with incoming connections and stops when idle. It
// is only valid when SourceAddr, TargetAddr, RunFunc, and StopFunc are all set.
type StartStopConfig struct {
	// SourceAddr is the address to listen for connections.
	SourceAddr net.Addr

	// TargetAddr is the address to forward all connections from SourceAddr.
	TargetAddr net.Addr

	// IdleDur is the duration to wait when idle before stopping the server or
	// procecss.
	IdleDur time.Duration

	// RunFunc defines the behavior to run a server or process. It should block
	// block until the server or process terminates.
	RunFunc func(ctx context.Context) error

	// StopFunc defines the behavior to stop a server or process started by
	// RunFunc. It should simply signal RunFunc to terminate gracefully and not
	// wait for RunFunc to return.
	StopFunc func(ctx context.Context) error

	// ErrorLog specifies an optional logger for unexpected behavior from
	// handling connections. If nil, logging is done via the log package's
	// standard logger.
	ErrorLog *log.Logger
}

func (ssc *StartStopConfig) logf(format string, v ...interface{}) {
	if ssc.ErrorLog != nil {
		ssc.ErrorLog.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

func (ssc *StartStopConfig) valid() error {
	if ssc.SourceAddr == nil {
		return ErrStartStopMissingField("SourceAddr")
	}
	if ssc.TargetAddr == nil {
		return ErrStartStopMissingField("TargetAddr")
	}
	if ssc.RunFunc == nil {
		return ErrStartStopMissingField("RunFunc")
	}
	if ssc.StopFunc == nil {
		return ErrStartStopMissingField("StopFunc")
	}
	return nil
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

func (ssc *StartStopConfig) serve(ctx context.Context, sc net.Conn) {
	defer sc.Close()

	// Wait for TargetAddr to open then dial it
	var d net.Dialer
	var tc net.Conn
	for tc == nil {
		tc, _ = d.DialContext(ctx, ssc.TargetAddr.Network(), ssc.TargetAddr.String())
		if ctx.Err() != nil {
			return
		}
	}
	defer tc.Close()

	// Establish bi-directional communication between tc and sc
	errCh := make(chan error)
	nDone := 0
	go func() {
		_, err := io.Copy(tc, sc)
		errCh <- err
	}()
	go func() {
		_, err := io.Copy(sc, tc)
		errCh <- err
	}()

	// Wait for the connection to close or return an error
	for {
		select {
		case err := <-errCh:
			nDone++
			if err != nil {
				ssc.logf("unexpected connection error: %s", err.Error())
			}
			if err != nil || nDone == 2 {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// Run runs a server or process that automatically starts with incoming
// connections and stops when idle.
func (ssc *StartStopConfig) Run(ctx context.Context) error {
	if err := ssc.valid(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Listen for connections on ssc.SourceAddr
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

	// Create a new (inactive) timer
	tIdle := time.NewTimer(ssc.IdleDur)
	tIdleOn := false
	if !tIdle.Stop() {
		<-tIdle.C
	}
	defer tIdle.Stop()

	// Handle incoming connections and start/stop functionality
	var sActive bool                 // Whether server is running
	var nConn int                    // Number of active connections
	sErrCh := make(chan error)       // Server error (can be nil; implies termination)
	cClosedCh := make(chan struct{}) // Per connection closed
	for {
		select {
		case c := <-connCh:
			// Stop the idle timer
			if tIdleOn {
				if !tIdle.Stop() {
					<-tIdle.C
				}
				tIdleOn = false
			}

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
				ssc.serve(ctx, c)
				cClosedCh <- struct{}{}
			}()
		case <-tIdle.C:
			// The timer should only trigger when the server is active
			if !sActive {
				panic("idle timer expired with an inactive server")
			}

			// Stop the server
			if err := ssc.StopFunc(ctx); err != nil {
				return err
			}
			<-sErrCh // Wait for ssc.RunFunc to return
			sActive = false
			tIdleOn = false
		case err := <-lErrCh:
			return err
		case err := <-sErrCh:
			if err != nil {
				return err
			}
			sActive = false

			// Stop the idle timer
			if tIdleOn {
				if !tIdle.Stop() {
					<-tIdle.C
				}
				tIdleOn = false
			}
		case <-cClosedCh:
			// The idle timer should not be running before any connection closes.
			if tIdleOn {
				panic("idle timer running or expired before a connection closed")
			}

			// Start the idle timer if there are no other connections
			nConn--
			if nConn == 0 {
				tIdle.Reset(ssc.IdleDur)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
