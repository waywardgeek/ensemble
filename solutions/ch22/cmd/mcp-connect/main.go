// Command mcp-connect joins an MCP client that speaks stdio to an MCP server
// listening on a TCP port: it copies stdin to the socket and the socket to
// stdout, and exits when either side closes.
//
// The ensemble agent started with --mcp-port exposes its GUI this way, so any
// MCP client that can launch a subprocess can see and drive the GUI. In an
// mcpServers configuration:
//
//	"ensemble-gui": {"command": "mcp-connect", "args": ["127.0.0.1:8095"]}
//
// A bare port means loopback. This is the whole transport: the wire on the
// port is the stdio wire, one JSON-RPC message per line, so there is nothing
// to translate.
package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Fprintln(os.Stderr, "usage: mcp-connect [HOST:]PORT")
		os.Exit(2)
	}
	addr := os.Args[1]
	if !strings.Contains(addr, ":") {
		addr = "127.0.0.1:" + addr
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// Stderr, never stdout: stdout belongs to the JSON-RPC stream, and a
		// stray line there is a parse error in the client.
		fmt.Fprintf(os.Stderr, "mcp-connect: %v (is ensemble running with --mcp-port?)\n", err)
		os.Exit(1)
	}
	done := make(chan struct{}, 2)
	go func() { io.Copy(conn, os.Stdin); done <- struct{}{} }()
	go func() { io.Copy(os.Stdout, conn); done <- struct{}{} }()
	<-done
	conn.Close()
}
