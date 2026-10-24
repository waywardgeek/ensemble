package tool

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type ptyChunks struct {
	chunks [][]byte
	final  error
}

func (r *ptyChunks) Read(p []byte) (int, error) {
	if len(r.chunks) == 0 {
		return 0, r.final
	}
	chunk := r.chunks[0]
	r.chunks = r.chunks[1:]
	n := copy(p, chunk)
	if len(r.chunks) == 0 {
		return n, r.final
	}
	return n, nil
}
func TestPTYNormalizationReadBoundaries(t *testing.T) {
	raw := []byte("α\r\nB\rC\r\r\nD\n終\r")
	want := []byte("α\nB\rC\r\nD\n終\r")
	for split := 0; split <= len(raw); split++ {
		var output bytes.Buffer
		reader := &ptyChunks{chunks: [][]byte{raw[:split], raw[split:]}, final: io.EOF}
		if err := copyPTY(&output, reader); err != nil || !bytes.Equal(output.Bytes(), want) {
			t.Fatalf("split %d: bytes=%q err=%v", split, output.Bytes(), err)
		}
	}
	var chunks [][]byte
	for i := range raw {
		chunks = append(chunks, raw[i:i+1])
	}
	readFailure := errors.New("PTY read failure")
	var output bytes.Buffer
	if err := copyPTY(&output, &ptyChunks{chunks: chunks, final: readFailure}); err != readFailure || !bytes.Equal(output.Bytes(), want) {
		t.Fatalf("single-byte UTF8/lone CR/error: bytes=%q err=%v", output.Bytes(), err)
	}
}
