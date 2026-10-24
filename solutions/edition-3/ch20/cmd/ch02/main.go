// Command ch02 uses the shared core in protocol, chat, render and dump modes.
package main

import (
	"context"
	"ensemble/internal/cli"
	"fmt"
	"io"
	"os"
	"os/signal"
)

func main() {
	ctx := context.Background()
	stop := func() {}
	if len(os.Args) != 2 || os.Args[1] != "chat" {
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt)
	}
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(ctx context.Context, args []string, in io.Reader, out, diagnostics io.Writer) error {
	return cli.Run(ctx, args, in, out, diagnostics)
}
