package eventlog

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"example.com/ensemble/internal/common"
)

// A streaming source keeps the input fixture out of the reader's allocation
// budget and reveals whether overflow drains an arbitrarily large source.
type spaces struct{ read int64 }

func (s *spaces) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = ' '
	}
	s.read += int64(len(p))
	return len(p), nil
}

func TestSkillRawRecordBound(t *testing.T) {
	fact := common.SkillTransition{Action: "initialize", Name: "base", Ceiling: []string{"load_skill", "unload_skill"}, State: common.SkillState{Primary: "base", Roots: []string{}, Active: []common.SkillActive{}, Available: []common.SkillOffer{}, Tools: []string{}, Retired: []common.SkillRetired{}}, Activated: []common.SkillActivation{}}
	encoded, err := json.Marshal(common.Event{Seq: 1, Type: "skills_initialized", Skills: &fact})
	if err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"\n", ""} {
		for _, excess := range []int{0, 1} {
			t.Run(strings.ReplaceAll(suffix, "\n", "LF")+string(rune('0'+excess)), func(t *testing.T) {
				padding := int64(common.SkillRecordLimit + excess - len(encoded) - len(suffix))
				input := io.MultiReader(strings.NewReader("{\"log_version\":1}\n"), io.LimitReader(&spaces{}, padding), strings.NewReader(string(encoded)+suffix))
				events, lines, err := Read(&owner{}, input)
				if excess == 0 {
					if err != nil || len(events) != 1 || lines[0] != 2 {
						t.Fatalf("exact raw bound: %v", err)
					}
				} else if err == nil || !strings.Contains(err.Error(), "line 2") {
					t.Fatalf("one over: %v", err)
				}
			})
		}
	}
	source := &spaces{}
	if _, err := readRecord(&owner{}, bufio.NewReader(source)); err == nil {
		t.Fatal("unbounded record accepted")
	}
	if source.read > common.SkillRecordLimit+4096 {
		t.Fatalf("drained excessive bytes: %d", source.read)
	}
}

func TestOrdinaryPhysicalBoundsRemain(t *testing.T) {
	for _, suffix := range []string{"\n", ""} {
		for _, size := range []int{16*1024*1024 - 1, 16 * 1024 * 1024, 16*1024*1024 + 1} {
			prefix := `{"log_version":1}`
			input := io.MultiReader(strings.NewReader(prefix), io.LimitReader(&spaces{}, int64(size-len(prefix)-len(suffix))), strings.NewReader(suffix))
			_, _, err := Read(&owner{}, input)
			want := size < 16*1024*1024 || size == 16*1024*1024 && suffix == "\n"
			if (err == nil) != want {
				t.Fatalf("ordinary bytes=%d framing=%q err=%v", size, suffix, err)
			}
		}
	}
}
