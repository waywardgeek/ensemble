package grade

// ch12Checks returns the check list for the MCP chapter.
func Ch12Checks(r Ch12Result) []Check {
	return []Check{
		ch12Handshake(r),
		ch12Discovery(r),
		ch12ToolCall(r),
		ch12Ephemeral(r),
		ch12Reverse(r),
		ch12WSTunnel(r),
		ch12Ch11Parity(r),
	}
}

func ch12Handshake(r Ch12Result) Check {
	c := Check{
		ID:     "mcp-handshake",
		Title:  "MCP initialize/initialized exchange over PipeTransport",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.HandshakeOK {
		c.failf("handshake: %s", r.HandshakeErr)
	}
	return c
}

func ch12Discovery(r Ch12Result) Check {
	c := Check{
		ID:     "tool-discovery",
		Title:  "tools/list populates agent registry with MCP tools",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.DiscoveryOK {
		c.failf("discovery: %s", r.DiscoveryErr)
	}
	return c
}

func ch12ToolCall(r Ch12Result) Check {
	c := Check{
		ID:     "mcp-tool-call",
		Title:  "MCP tool call via bridge returns result",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.ToolCallOK {
		c.failf("tool call: %s", r.ToolCallErr)
	}
	return c
}

func ch12Ephemeral(r Ch12Result) Check {
	c := Check{
		ID:     "ephemeral-round",
		Title:  "Ephemeral tool auto-called each round; result in Ephemera",
		Points: 20,
		Earned: 20,
		Passed: true,
	}
	if !r.EphemeralOK {
		c.failf("ephemeral: %s", r.EphemeralErr)
	}
	return c
}

func ch12Reverse(r Ch12Result) Check {
	c := Check{
		ID:     "reverse-call",
		Title:  "MCP server calls agent tool via reverse handler",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.ReverseOK {
		c.failf("reverse: %s", r.ReverseErr)
	}
	return c
}

func ch12WSTunnel(r Ch12Result) Check {
	c := Check{
		ID:     "ws-tunnel",
		Title:  "JSON-RPC messages routed through WebSocket hub",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.WSTunnelOK {
		c.failf("ws-tunnel: %s", r.WSTunnelErr)
	}
	return c
}

func ch12Ch11Parity(r Ch12Result) Check {
	c := Check{
		ID:     "ch11-parity",
		Title:  "All ch11 behavior preserved",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.Ch11Parity {
		c.failf("ch11 parity: %s", r.Ch11ParityErr)
	}
	return c
}
