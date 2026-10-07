package main

import (
	"os"

	"example.com/ensemble"
)

func isTerminal(owner ensemble.ClientOwner, stream any) bool {
	file, ok := stream.(*os.File)
	return ok && terminalFD(owner, file.Fd())
}
