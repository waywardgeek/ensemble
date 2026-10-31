package eventlog

import "io"

// Review-only wrapper; the real descriptor and close lifetime remain in use.
func (l *Log) Ch10ReviewWrapWriter(wrap func(io.WriteCloser) io.WriteCloser) {
	l.writer = wrap(l.writer)
}
