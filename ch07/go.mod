module github.com/waywardgeek/ensemble/ch07

go 1.25

require github.com/waywardgeek/ensemble/agent v0.0.0

require (
	github.com/creack/pty v1.1.24 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
)

replace github.com/waywardgeek/ensemble/agent => ../agent
