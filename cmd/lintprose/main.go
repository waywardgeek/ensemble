// lintprose counts the prose moves that book/voice.md §5 budgets, per file,
// and fails when a hard ceiling is exceeded. It measures density, not
// quality: it cannot tell a good short closer from a bad one, only that four
// in five paragraphs end the same way.
//
// Usage: go run ./cmd/lintprose [-v] book/preface.md book/chapter-*.md
//
// Exit 0: all hard ceilings met. Exit 1: at least one hard failure.
// Soft ceilings print WARN and never change the exit code.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Budgets from voice.md §5. Change the document first, then this table.
const (
	minWords          = 4000 // chapters only; the preface is exempt
	maxWords          = 7500
	maxConfessions    = 3
	maxDefenses       = 1
	negationPer       = 500 // soft: one per this many words
	maxInventory      = 3
	shortCloserWords  = 8
	shortCloserRatio  = 1.0 / 3.0 // soft
	maxPointerClosers = 2
	maxSuperlatives   = 2
	maxPretend        = 1
	personGapWords    = 1200 // soft: mechanism with nobody on the page
	headingRunWarn    = 3    // soft: consecutive headings of one shape
)

var (
	confessionRe = regexp.MustCompile(`(?i)\b(I believed|I had this wrong|I wrote it that way|an earlier draft|the first draft of|shipped a bug|I went into this)\b`)
	defenseRe    = regexp.MustCompile(`(?i)\b(looks like over-|feels like over-|looks like a violation)`)
	negationRe   = regexp.MustCompile(`, not |, never |(^|[.!?] )Not `)
	inventoryRe  = regexp.MustCompile(`^(Two|Three|Four|Five|Six|Seven|Eight|Nine|Ten|Twelve)\b`)
	pointerRe    = regexp.MustCompile(`^(That is|That sentence|That number|That habit|That shape|That notation|Read that|Look at|Hold on to|Here is the)\b`)
	superlatRe   = regexp.MustCompile(`(?i)\bthe (most \w+|only|cheapest|purest|single most|whole)\b[^.]{0,60}?\b(in|of) (the|this) (chapter|book)\b`)
	pretendRe    = regexp.MustCompile(`(?i)I won'?t pretend otherwise`)
	emDashRe     = regexp.MustCompile("\u2014")
	// "delve" alone would flag the Go debugger (go-delve/delve); the crutch
	// is the verb. "not just" alone flags ordinary usage; the crutch is the
	// "not just X, (it's|but) Y" frame.
	crutchRe     = regexp.MustCompile(`(?i)\bdelv(e|es|ed|ing) into\b|\btapestry\b|\btestament to\b|\bnot just \w+[^.]{0,40}\b(but|it'?s)\b|\bit'?s not \w+[^.]{0,40}, it'?s\b`)
	personRe     = regexp.MustCompile(`\bBill\b|\bI\b|\bmy\b|\bme\b`)
	negHeadingRe = regexp.MustCompile(`, not |\bis not\b|, never |\bis a\b|\bis the\b`)
	sentenceEnd  = regexp.MustCompile(`[.!?]["')\]]*\s+`)
)

type para struct {
	text string
	line int
}

type doc struct {
	paras    []para
	headings []string
	emDashes int
	words    int
}

func parse(path string) (*doc, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := &doc{}
	var cur []string
	curLine := 0
	inCode := false
	flush := func() {
		if len(cur) == 0 {
			return
		}
		text := strings.Join(cur, " ")
		cur = nil
		text = strings.TrimSpace(text)
		if text == "" || strings.HasPrefix(text, "[") { // [VERIFY ...] notes
			return
		}
		d.paras = append(d.paras, para{text: text, line: curLine})
		d.words += len(strings.Fields(text))
	}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	n := 0
	for sc.Scan() {
		n++
		line := sc.Text()
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			flush()
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		d.emDashes += len(emDashRe.FindAllString(line, -1))
		switch {
		case trim == "":
			flush()
		case strings.HasPrefix(trim, "|"):
			flush()
		case strings.HasPrefix(trim, "#"):
			flush()
			d.headings = append(d.headings, strings.TrimLeft(trim, "# "))
		default:
			trim = strings.TrimPrefix(trim, "> ")
			trim = strings.TrimPrefix(trim, ">")
			if len(cur) == 0 {
				curLine = n
			}
			cur = append(cur, trim)
		}
	}
	flush()
	return d, sc.Err()
}

// lastSentence returns the final sentence of a paragraph, with list markers
// and emphasis stripped so word counts are honest.
func lastSentence(p string) string {
	p = strings.TrimRight(p, " \t")
	idx := sentenceEnd.FindAllStringIndex(p, -1)
	last := p
	if len(idx) > 0 {
		last = p[idx[len(idx)-1][1]:]
	}
	return strings.Trim(last, "*_ ")
}

func firstSentence(p string) string {
	p = strings.TrimLeft(p, "-*0123456789. ")
	p = strings.Trim(p, "*_")
	if m := sentenceEnd.FindStringIndex(p); m != nil {
		return p[:m[0]]
	}
	return p
}

type result struct {
	name   string
	count  int
	limit  string
	hard   bool
	failed bool
}

func lint(path string, verbose bool) ([]result, error) {
	d, err := parse(path)
	if err != nil {
		return nil, err
	}
	all := make([]string, 0, len(d.paras))
	for _, p := range d.paras {
		all = append(all, p.text)
	}
	body := strings.Join(all, "\n")

	var res []result
	add := func(name string, count int, limit string, hard, failed bool) {
		res = append(res, result{name, count, limit, hard, failed})
	}

	isChapter := strings.HasPrefix(filepath.Base(path), "chapter-")
	if isChapter {
		add("words", d.words, fmt.Sprintf("%d..%d", minWords, maxWords), true,
			d.words < minWords || d.words > maxWords)
	} else {
		add("words", d.words, "exempt", false, false)
	}

	c := len(confessionRe.FindAllString(body, -1))
	add("confessions", c, fmt.Sprint("<=", maxConfessions), true, c > maxConfessions)

	c = len(defenseRe.FindAllString(body, -1))
	add("self-defense", c, fmt.Sprint("<=", maxDefenses), true, c > maxDefenses)

	c = len(negationRe.FindAllString(body, -1))
	negLimit := d.words / negationPer
	add("negation forms", c, fmt.Sprintf("<=%d (1/%dw)", negLimit, negationPer), false, c > negLimit)

	inv, short, pointer := 0, 0, 0
	var shortList []string
	for _, p := range d.paras {
		fs := firstSentence(p.text)
		if inventoryRe.MatchString(fs) && len(strings.Fields(fs)) <= 6 {
			inv++
			if verbose {
				fmt.Printf("    inventory  L%d: %s\n", p.line, fs)
			}
		}
		ls := lastSentence(p.text)
		if n := len(strings.Fields(ls)); n > 0 && n <= shortCloserWords {
			short++
			shortList = append(shortList, fmt.Sprintf("L%d: %s", p.line, ls))
		}
		if pointerRe.MatchString(ls) {
			pointer++
			if verbose {
				fmt.Printf("    pointer    L%d: %s\n", p.line, ls)
			}
		}
	}
	if verbose {
		for _, s := range shortList {
			fmt.Printf("    short      %s\n", s)
		}
	}
	add("inventory openers", inv, fmt.Sprint("<=", maxInventory), true, inv > maxInventory)
	shortLimit := int(float64(len(d.paras)) * shortCloserRatio)
	add("short closers", short, fmt.Sprintf("<=%d of %d paras", shortLimit, len(d.paras)), false, short > shortLimit)
	add("pointer closers", pointer, fmt.Sprint("<=", maxPointerClosers), true, pointer > maxPointerClosers)

	c = len(superlatRe.FindAllString(body, -1))
	add("scope superlatives", c, fmt.Sprint("<=", maxSuperlatives), true, c > maxSuperlatives)

	c = len(pretendRe.FindAllString(body, -1))
	add("won't pretend", c, fmt.Sprint("<=", maxPretend), true, c > maxPretend)

	add("em-dashes", d.emDashes, "0", true, d.emDashes > 0)

	c = len(crutchRe.FindAllString(body, -1))
	if verbose && c > 0 {
		for _, m := range crutchRe.FindAllString(body, -1) {
			fmt.Printf("    crutch     %q\n", m)
		}
	}
	add("LLM crutches", c, "0", true, c > 0)

	// Person gap: longest run of words with nobody on the page.
	gap, maxGap, gapStart, worstStart := 0, 0, 0, 0
	for _, p := range d.paras {
		if personRe.MatchString(p.text) {
			gap = 0
			gapStart = p.line
			continue
		}
		gap += len(strings.Fields(p.text))
		if gap > maxGap {
			maxGap, worstStart = gap, gapStart
		}
	}
	add(fmt.Sprintf("person gap (after L%d)", worstStart), maxGap, fmt.Sprint("<=", personGapWords), false, maxGap > personGapWords)

	// Heading runs of one shape.
	run, maxRun := 0, 0
	for _, h := range d.headings {
		if negHeadingRe.MatchString(h) {
			run++
			if run > maxRun {
				maxRun = run
			}
		} else {
			run = 0
		}
	}
	add("heading shape run", maxRun, fmt.Sprint("<", headingRunWarn), false, maxRun >= headingRunWarn)

	return res, nil
}

func main() {
	verbose := flag.Bool("v", false, "list every counted instance with line numbers")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: lintprose [-v] FILE.md ...")
		os.Exit(2)
	}
	exit := 0
	for _, path := range flag.Args() {
		fmt.Printf("%s\n", path)
		res, err := lint(path, *verbose)
		if err != nil {
			fmt.Printf("  error: %v\n", err)
			exit = 2
			continue
		}
		for _, r := range res {
			status := "ok  "
			if r.failed {
				if r.hard {
					status = "FAIL"
					exit = 1
				} else {
					status = "WARN"
				}
			}
			fmt.Printf("  %s %-28s %6d   %s\n", status, r.name, r.count, r.limit)
		}
	}
	os.Exit(exit)
}
