package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

var errPeerEOF = errors.New("interactive peer closed stdin")

// One owner watches stdin independently of executing commands or listener Ask.
// Queuing input instead of blocking on a command consumer lets EOF cancel even
// a pending sleep/wait. Normal one-shot mode deliberately has no peer owner.
type peerInput struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	input  io.ReadCloser
	mu     sync.Mutex
	lines  []string
	wake   chan struct{}
	done   chan struct{}
}

func watchPeer(parent context.Context, input *os.File, jsonl bool) (*peerInput, error) {
	// Inherited stdin may be a blocking descriptor: Close alone cannot wake
	// its pending read. A private nonblocking duplicate lets os.File use Go's
	// poller, so this owner can stop the scanner even while stdin remains open.
	fd, err := syscall.Dup(int(input.Fd()))
	if err != nil {
		return nil, fmt.Errorf("duplicate peer stdin: %w", err)
	}
	syscall.CloseOnExec(fd)
	if err = syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("poll peer stdin: %w", err)
	}
	reader := os.NewFile(uintptr(fd), "peer-stdin")
	ctx, cancel := context.WithCancelCause(parent)
	p := &peerInput{ctx: ctx, cancel: cancel, input: reader, wake: make(chan struct{}, 1), done: make(chan struct{})}
	go func() {
		defer close(p.done)
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 4096), 1<<20)
		for scanner.Scan() {
			if ctx.Err() != nil {
				return
			}
			if jsonl {
				p.mu.Lock()
				p.lines = append(p.lines, scanner.Text())
				p.mu.Unlock()
				select {
				case p.wake <- struct{}{}:
				default:
				}
			}
		}
		if err := scanner.Err(); err != nil {
			cancel(fmt.Errorf("peer stdin: %w", err))
		} else {
			cancel(errPeerEOF)
		}
	}()
	return p, nil
}
func (p *peerInput) next() (string, error) {
	for {
		if err := context.Cause(p.ctx); err != nil {
			return "", err
		}
		p.mu.Lock()
		if len(p.lines) > 0 {
			line := p.lines[0]
			p.lines[0] = ""
			p.lines = p.lines[1:]
			p.mu.Unlock()
			return line, nil
		}
		p.mu.Unlock()
		select {
		case <-p.ctx.Done():
			return "", context.Cause(p.ctx)
		case <-p.wake:
		}
	}
}
func (p *peerInput) close() error {
	p.cancel(context.Canceled)
	err := p.input.Close()
	<-p.done
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
		return nil
	}
	return err
}
