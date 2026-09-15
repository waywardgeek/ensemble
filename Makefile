.PHONY: test grade grade2 grade3 grade-dir fake fake-serve vet fmt lint-prose

# Count the prose moves budgeted in book/voice.md §5. Hard ceilings fail;
# soft ones warn. Add -v via LINTFLAGS=-v to list every counted instance.
lint-prose:
	go run ./cmd/lintprose $(LINTFLAGS) book/preface.md book/chapter-[0-9][0-9].md

test: vet
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

# Grade the reference solutions.
grade:
	go run ./cmd/grade -ch 1 ./solutions/ch01

grade2:
	go run ./cmd/grade -ch 2 ./solutions/ch02

grade3:
	go run ./cmd/grade -ch 3 ./solutions/ch03

# Grade an arbitrary submission:
#   make grade-dir DIR=/path/to/submission        (chapter 1)
#   make grade-dir CH=3 DIR=/path/to/submission   (chapter 3)
CH ?= 1
grade-dir:
	go run ./cmd/grade -ch $(CH) $(DIR)

# Run a reference solution against the fake vendor and watch the loop turn.
#   make fake                           chapter 3 REPL, Anthropic dialect
#   make fake VENDOR=gemini             same loop, Gemini dialect
#   make fake FAKE_CH=2 VENDOR=openai   chapter 2, OpenAI dialect
# Or serve the fake alone and paste the env block into your own agent's shell:
#   make fake-serve VENDOR=openai
VENDOR ?= anthropic
FAKE_CH ?= 3
fake:
	go run ./cmd/fakevendor -ch $(FAKE_CH) -vendor $(VENDOR) chat

fake-serve:
	go run ./cmd/fakevendor -vendor $(VENDOR)
