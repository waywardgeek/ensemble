.PHONY: test grade grade2 grade3 grade-dir fake fake-serve vet fmt lint-prose book

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

# Grade the reference solutions — all chapters grade from the student's
# directory (solutions/chNN for frozen snapshots, ./agent for the live code).
# The grade-dir target lets you point at any directory:
#   make grade-dir CH=8 DIR=./agent
grade:
	go run ./cmd/grade -ch 1 ./solutions/ch01

grade2:
	go run ./cmd/grade -ch 2 ./solutions/ch02

grade3:
	go run ./cmd/grade -ch 3 ./solutions/ch03

grade5:
	go run ./cmd/grade -ch 5 ./agent

# Note: ch6 full grading requires the frozen solution (old repo-root shape).
# From ch7 onward, ch6 features are tested via parity checks.
# Use: make grade-dir CH=6 DIR=./solutions/ch06/agent
# (scores 45/100 — exercise-specific multi-agent checks need the old shape)

grade7:
	go run ./cmd/grade -ch 7 ./agent

grade8:
	go run ./cmd/grade -ch 8 ./agent

grade9:
	go run ./cmd/grade -ch 9 ./agent

grade-dir:
	go run ./cmd/grade -ch $(CH) $(DIR)

grade-ch07: grade7

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

# Assemble the full book into a single readable markdown file.
book:
	scripts/assemble-book.sh > book/the-self-wielding-agent.md
	@wc -w < book/the-self-wielding-agent.md | xargs -I{} echo "book/the-self-wielding-agent.md: {} words"
