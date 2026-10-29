// This local test peer imports a byte-for-byte copy of the existing Go Gateway
// wrapper. It is not an alternative implementation or a production service.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"noiseinterop/gateway"
	"os"
)

type request struct {
	Operation  string `json:"operation"`
	Role       string `json:"role"`
	Data       []byte `json:"data"`
	Associated []byte `json:"associated"`
	Peer       []byte `json:"peer"`
	Prologue   []byte `json:"prologue"`
}
type reply struct {
	Data     []byte `json:"data"`
	Error    string `json:"error"`
	Complete bool   `json:"complete"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	output := json.NewEncoder(os.Stdout)
	var session *gateway.NoiseKK
	for scanner.Scan() {
		var r request
		var result reply
		err := json.Unmarshal(scanner.Bytes(), &r)
		if err == nil {
			if r.Operation == "start" {
				if session != nil {
					err = fmt.Errorf("session already started")
				} else {
					var key gateway.Keypair
					key, err = gateway.GenerateKeypair()
					if err == nil {
						session, err = gateway.NewNoiseKK(r.Role, key, r.Peer, r.Prologue)
						result.Data = key.Public
					}
				}
			} else if session == nil {
				err = fmt.Errorf("session not started")
			} else {
				switch r.Operation {
				case "write":
					result.Data, err = session.WriteHandshakeMessage(r.Data)
				case "read":
					result.Data, err = session.ReadHandshakeMessage(r.Data)
				case "encrypt":
					result.Data, err = session.Encrypt(r.Data, r.Associated)
				case "decrypt":
					result.Data, err = session.Decrypt(r.Data, r.Associated)
				default:
					err = fmt.Errorf("unknown operation")
				}
			}
		}
		if err != nil {
			result.Error = err.Error()
		}
		if session != nil {
			result.Complete = session.IsHandshakeComplete()
		}
		if err := output.Encode(result); err != nil {
			panic(err)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
