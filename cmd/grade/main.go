// Command grade runs a student submission against the course auto-grader.
//
//	grade [-ch N] [-json] [DIR]
//
// DIR is the student's code directory. For ch1-4 it's a flat package; for
// ch5+ it's a library tree. With no path, it grades the reference solution
// for the chosen chapter.
//
// Exit status: 0 pass, 1 fail, 2 the grader itself could not run.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/waywardgeek/ensemble/internal/grade"
)

func main() {
	var (
		asJSON  = flag.Bool("json", false, "emit the report as JSON")
		chapter = flag.Int("ch", 1, "which chapter's exercise to grade")
	)
	flag.Parse()

	path := flag.Arg(0)
	if path == "" {
		path = fmt.Sprintf("./solutions/ch%02d", *chapter)
	}

	var report grade.Report
	switch *chapter {
	case 1:
		bin, cleanup, err := grade.Build(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		defer cleanup()
		res, err := grade.Run(bin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewReport(grade.Evaluate(res), res.Stderr)
	case 2:
		bin, cleanup, err := grade.Build(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		defer cleanup()
		res, err := grade.Ch2Run(bin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewTitledReport(
			"Chapter 2 — One Log, Three Vendors", grade.Ch2Evaluate(res), res.Stderr)
	case 3:
		bin, cleanup, err := grade.Build(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		defer cleanup()
		res, err := grade.Ch3Run(bin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewTitledReport(
			"Chapter 3 — Six Tools: Ninety-Two Percent of an AI Coding Agent",
			grade.Ch3Evaluate(res), "")
	case 4:
		bin, cleanup, err := grade.Build(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		defer cleanup()
		res, err := grade.Ch4Run(bin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewTitledReport(
			"Chapter 4 — Jobs: Containment, Not Cancellation",
			grade.Ch4Evaluate(res), res.HelpersErr)
	case 5:
		res, err := grade.Ch5Run(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewTitledReport(
			"Chapter 5 — The Big Refactor",
			grade.Ch5Evaluate(res), res.HelpersErr)
	case 6:
		res, err := grade.Ch6Run(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewTitledReport(
			"Chapter 6 — Two Seams and a Loop",
			grade.Ch6Evaluate(res), res.HelpersErr)
	case 7:
		res, err := grade.Ch7Run(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewTitledReport(
			"Chapter 7 — Streaming",
			grade.Ch7Evaluate(res), res.HelpersErr)
	case 8:
		res, err := grade.Ch8Run(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewTitledReport(
			"Chapter 8 — Everything Is an Artifact",
			grade.Ch8Evaluate(res), res.HelpersErr)
	case 9:
		res, err := grade.Ch9Run(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewTitledReport(
			"Chapter 9 — Build Your Dream GUI",
			grade.Ch9Evaluate(res), res.HelpersErr)
	case 10:
		res, err := grade.Ch10Run(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			os.Exit(2)
		}
		report = grade.NewTitledReport(
			"Chapter 10 — Skills",
			grade.Ch10Evaluate(res), res.HelpersErr)
	default:
		fmt.Fprintf(os.Stderr, "grader: no grader for chapter %d yet\n", *chapter)
		os.Exit(2)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(report)
	} else {
		report.WriteText(os.Stdout)
	}
	if !report.Passed {
		os.Exit(1)
	}
}
