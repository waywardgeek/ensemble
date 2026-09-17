// Command hanoi solves the Tower of Hanoi puzzle.
//
// It moves a stack of n disks from a source peg to a destination peg,
// never placing a larger disk on top of a smaller one, and prints each
// move it makes.
//
// Usage:
//
//	hanoi [-n disks] [-show]
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// A Move is a single transfer of the top disk from one peg to another.
type Move struct {
	Disk int // disk size, 1 is smallest
	From int // peg index, 0-based
	To   int // peg index, 0-based
}

// Solve returns the sequence of moves that transfers n disks from peg
// from to peg to, using peg via as scratch space. The result has
// 2^n - 1 moves, which is the fewest possible.
func Solve(n, from, to, via int) []Move {
	if n <= 0 {
		return nil
	}
	moves := make([]Move, 0, 1<<n-1)
	var rec func(n, from, to, via int)
	rec = func(n, from, to, via int) {
		if n == 0 {
			return
		}
		// Move the n-1 disks above the biggest one out of the way,
		// shift the biggest disk, then pile the rest back on top.
		rec(n-1, from, via, to)
		moves = append(moves, Move{Disk: n, From: from, To: to})
		rec(n-1, via, to, from)
	}
	rec(n, from, to, via)
	return moves
}

// Pegs holds the disks on each of the three pegs, bottom disk first.
type Pegs [3][]int

// NewPegs returns three pegs with n disks stacked on peg start, largest
// at the bottom.
func NewPegs(n, start int) *Pegs {
	p := &Pegs{}
	for disk := n; disk >= 1; disk-- {
		p[start] = append(p[start], disk)
	}
	return p
}

// Apply performs m, reporting an error if the move breaks the rules.
func (p *Pegs) Apply(m Move) error {
	src := p[m.From]
	if len(src) == 0 {
		return fmt.Errorf("peg %c is empty", pegName(m.From))
	}
	top := src[len(src)-1]
	if top != m.Disk {
		return fmt.Errorf("disk %d is not on top of peg %c", m.Disk, pegName(m.From))
	}
	if dst := p[m.To]; len(dst) > 0 && dst[len(dst)-1] < m.Disk {
		return fmt.Errorf("cannot put disk %d on smaller disk %d", m.Disk, dst[len(dst)-1])
	}
	p[m.From] = src[:len(src)-1]
	p[m.To] = append(p[m.To], m.Disk)
	return nil
}

func pegName(i int) rune {
	return rune('A' + i)
}

// String draws the pegs as ASCII art, tallest peg first.
func (p *Pegs) String() string {
	n := 0
	for _, peg := range p {
		n += len(peg)
	}
	width := 2*n + 1 // widest disk
	var b strings.Builder
	for row := n - 1; row >= 0; row-- {
		for i, peg := range p {
			if i > 0 {
				b.WriteByte(' ')
			}
			if row < len(peg) {
				b.WriteString(center(disk(peg[row]), width))
			} else {
				b.WriteString(center("|", width))
			}
		}
		b.WriteByte('\n')
	}
	for i := range p {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(center(string(pegName(i)), width))
	}
	b.WriteByte('\n')
	return b.String()
}

func disk(size int) string {
	return strings.Repeat("=", 2*size+1)
}

func center(s string, width int) string {
	pad := width - len(s)
	if pad <= 0 {
		return s
	}
	left := pad / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
}

func main() {
	n := flag.Int("n", 3, "number of disks")
	show := flag.Bool("show", false, "draw the pegs after every move")
	flag.Parse()

	if *n < 1 {
		fmt.Fprintf(os.Stderr, "hanoi: -n must be at least 1, got %d\n", *n)
		os.Exit(2)
	}

	const (
		src = 0
		dst = 2
		via = 1
	)

	pegs := NewPegs(*n, src)
	fmt.Printf("Tower of Hanoi with %d disks: %c -> %c\n", *n, pegName(src), pegName(dst))
	if *show {
		fmt.Printf("\n%s", pegs)
	}

	moves := Solve(*n, src, dst, via)
	for i, m := range moves {
		if err := pegs.Apply(m); err != nil {
			// Solve is correct, so this means a bug, not bad input.
			fmt.Fprintf(os.Stderr, "hanoi: illegal move %d: %v\n", i+1, err)
			os.Exit(1)
		}
		fmt.Printf("Move %d: disk %d from %c to %c\n", i+1, m.Disk, pegName(m.From), pegName(m.To))
		if *show {
			fmt.Printf("\n%s", pegs)
		}
	}
	fmt.Printf("Solved in %d moves.\n", len(moves))
}
