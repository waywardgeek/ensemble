package ws

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// TTSLogger records what entered the speech channel, one JSON object per line:
//
//	{"ms":412,"kind":"utterance","seq":1,"text":"Reading the file now."}
//	{"ms":1183,"kind":"utterance","seq":2,"text":"code block.","raw":"```go\nx := 1\n```"}
//	{"ms":2011,"kind":"pause"}
//	{"ms":4402,"kind":"resume"}
//
// Speech is the one output channel that cannot be scrolled back through. Every
// other artifact this system produces leaves something on screen to re-read;
// an utterance evaporates as it is spoken. That is why defects survived in it
// for so long, including two that a daily listener heard and filtered out as
// noise. This log turns an unrewindable channel into an inspectable one.
//
// The record is made client-side, because the decision about what is speakable
// is made client-side. The server cannot derive these lines; it can only be
// told them.
type TTSLogger struct {
	mu    sync.Mutex
	file  *os.File
	start time.Time
}

// TTSEvent is one line of the speech log.
//
// Timestamps are milliseconds since the log opened rather than wall clock. The
// questions this log answers are all about intervals: did two sentences arrive
// as one utterance or two, how long was the gate closed, did the queue stall.
// A relative number answers those by subtraction and keeps lines short enough
// to listen to, which matters when the person debugging the speech channel
// reads by speech.
//
// Empty fields are omitted so a gate transition is one short line rather than
// a row of empty strings.
type TTSEvent struct {
	Ms   int64  `json:"ms"`
	Kind string `json:"kind"`
	Seq  int    `json:"seq,omitempty"`
	Text string `json:"text,omitempty"`

	// The text before normalisation, present only when normalisation changed
	// it. This is the field that makes filter bugs visible: it shows what the
	// listener would have heard next to what they actually heard.
	Raw string `json:"raw,omitempty"`
}

// NewTTSLogger opens the speech log for writing. An empty path silently
// disables logging, matching GuiLogger.
func NewTTSLogger(path string) *TTSLogger {
	if path == "" {
		return &TTSLogger{}
	}
	f, err := os.Create(path)
	if err != nil {
		return &TTSLogger{}
	}
	return &TTSLogger{file: f, start: time.Now()}
}

// Log appends one event. Kind, Seq, Text and Raw come from the client; Ms is
// stamped here so every line shares one clock.
func (t *TTSLogger) Log(ev TTSEvent) {
	if t == nil || t.file == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	ev.Ms = time.Since(t.start).Milliseconds()
	line, err := json.Marshal(ev)
	if err != nil {
		return
	}
	t.file.Write(append(line, '\n'))
}

// Close closes the underlying file.
func (t *TTSLogger) Close() {
	if t == nil || t.file == nil {
		return
	}
	t.file.Close()
}
