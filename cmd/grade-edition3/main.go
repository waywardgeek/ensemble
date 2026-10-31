// Command grade-edition3 runs the supported Edition 3 chapter gates.
// Invoke it from the repository root with an explicit submission path.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/grade"
	"github.com/waywardgeek/ensemble/internal/gradeedition2"
	"github.com/waywardgeek/ensemble/internal/gradeedition3"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// Returning the status lets temporary build cleanup finish before os.Exit.
func run(args []string) int {
	flags := flag.NewFlagSet("grade-edition3", flag.ContinueOnError)
	chapter := flags.Int("ch", 1, "Edition 3 chapter to grade (1, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 18, 19, 20 or 21)")
	asJSON := flags.Bool("json", false, "emit the report as JSON")
	executableOnly := flags.Bool("executable-only", false, "Chapter 5 public CLI replay/media and non-disruptive hint checks")
	core := flags.String("core", "", "Chapter 4/5/6/7/8/9/10/11/12/13 core source module directory")
	consumer := flags.String("consumer", "", "Chapter 4/5/6/7/8/9/10/11/12/13 external consumer module directory")
	consumerCase := flags.String("consumer-case", "", "Chapter 4/5/6/7/8/9/10/11/12/13 JSON case file (default: shout example)")
	workflow := flags.String("workflow", "", "Chapter 5/6/7/8/9/10/11/12/13 workflow module directory")
	mediaModels := flags.String("media-models", "", "Chapter 5/6/7/8/9/10/11/12/13 media-model mapping JSON")
	streamExercise := flags.String("stream-exercise", "", "Chapter 6/7/8/9/10/11/12/13 stream-shape module directory")
	streamModel := flags.String("stream-model", "", "Chapter 6/7/8/9/10/11/12/13 declared known streaming model (required)")
	gui := flags.String("gui", "", "Chapter 7/8/9/10/11/12/13 separate GUI module directory")
	guiModel := flags.String("gui-model", "", "Chapter 7/8/9/10/11/12/13 declared GUI model (required)")
	cacheCurrent := flags.String("cache-current", "cr/cache-current.txt", "Chapter 15 current request artifact path in scratch launch directory")
	cachePrior := flags.String("cache-prior", "cr/cache-prior.txt", "Chapter 15 previous request artifact path in scratch launch directory")
	cachePrices := flags.String("cache-prices", "", "Chapter 15 required input,write,read,output rates per million tokens")
	statusPrices := flags.String("status-prices", "", "Chapter 18 required Sonnet 4.6 input,write,read,output rates per million")
	statusSecond := flags.String("status-second-prices", "", "Chapter 18 optional Opus 4.6 rates; absent selects an unpriced course model")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	chapter4Flags := false
	chapter5Flags := false
	chapter6Flags := false
	chapter7Flags := false
	chapter15Flags := false
	chapter18Flags := false
	set := map[string]bool{}
	flags.Visit(func(f *flag.Flag) {
		set[f.Name] = true
		switch f.Name {
		case "core", "consumer", "consumer-case":
			chapter4Flags = true
		case "workflow", "media-models":
			chapter5Flags = true
		case "stream-exercise", "stream-model":
			chapter6Flags = true
		case "gui", "gui-model":
			chapter7Flags = true
		case "cache-current", "cache-prior", "cache-prices":
			chapter15Flags = true
		case "status-prices", "status-second-prices":
			chapter18Flags = true
		}
	})
	if set["executable-only"] {
		if !*executableOnly || *chapter != 5 || flags.NArg() != 1 || strings.TrimSpace(flags.Arg(0)) == "" {
			fmt.Fprintln(os.Stderr, "-executable-only requires -ch 5 and one CLI_TARGET")
			return 2
		}
		for name := range set {
			if name != "ch" && name != "json" && name != "executable-only" && name != "media-models" {
				fmt.Fprintln(os.Stderr, "Chapter 5 -executable-only does not accept -"+name)
				return 2
			}
		}
		models, err := gradeedition2.ReadMediaModels(*mediaModels)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		bin, cleanup, err := grade.Build(flags.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		defer cleanup()
		report := gradeedition3.Chapter5Executable(bin, models)
		if *asJSON {
			if err := report.WriteJSON(os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 2
			}
		} else {
			report.WriteText(os.Stdout)
		}
		if !report.Passed {
			return 1
		}
		return 0
	}
	// Chapters 14-16 directly grade the GUI target; parity artifact flags do not apply.
	if (*chapter == 14 || *chapter == 15 || *chapter == 16) && (chapter4Flags || chapter5Flags || chapter6Flags || chapter7Flags) {
		fmt.Fprintln(os.Stderr, "Chapters 14/15/16 take a direct GUI_TARGET; omit CLI/parity artifact flags")
		return 2
	}
	var prices [4]float64
	if chapter15Flags && *chapter != 15 {
		fmt.Fprintln(os.Stderr, "cache artifact and price flags apply only to Chapter 15")
		return 2
	}
	if *chapter == 15 {
		var err error
		prices, err = gradeedition2.ParseCachePrices(*cachePrices)
		if err != nil || strings.TrimSpace(*cacheCurrent) == "" || strings.TrimSpace(*cachePrior) == "" {
			fmt.Fprintln(os.Stderr, "Chapter 15 requires -cache-prices input,write,read,output and nonblank artifact paths")
			return 2
		}
	}
	var statusRates [4]float64
	var secondRates *[4]float64
	if chapter18Flags && *chapter != 18 {
		fmt.Fprintln(os.Stderr, "status price flags apply only to Chapter 18")
		return 2
	}
	if *chapter == 18 {
		for _, name := range []string{"consumer", "consumer-case", "workflow", "media-models", "stream-exercise", "stream-model", "gui-model"} {
			if set[name] {
				fmt.Fprintln(os.Stderr, "Chapter 18 does not use -"+name)
				return 2
			}
		}
		info, err := os.Stat(filepath.Join(*core, "go.mod"))
		if strings.TrimSpace(*core) == "" || err != nil || !info.Mode().IsRegular() {
			fmt.Fprintln(os.Stderr, "Chapter 18 requires -core with an importable module")
			return 2
		}
		if strings.TrimSpace(*gui) == "" {
			fmt.Fprintln(os.Stderr, "Chapter 18 requires -gui GUI_TARGET")
			return 2
		}
		if _, err := os.Stat(*gui); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		statusRates, err = gradeedition2.ParseCachePrices(*statusPrices)
		if err != nil {
			fmt.Fprintln(os.Stderr, "-status-prices:", err)
			return 2
		}
		if set["status-second-prices"] {
			second, err := gradeedition2.ParseCachePrices(*statusSecond)
			if err != nil {
				fmt.Fprintln(os.Stderr, "-status-second-prices:", err)
				return 2
			}
			secondRates = &second
		}
	}
	if *chapter == 20 {
		if info, err := os.Stat(filepath.Join(*core, "go.mod")); err != nil || info.IsDir() {
			fmt.Fprintln(os.Stderr, "Chapter 20 requires -core to name a Go module directory")
			return 2
		}
		if *core == "" || *gui == "" {
			fmt.Fprintln(os.Stderr, "Chapter 20 requires -core CORE_DIR and -gui GUI_TARGET plus a positional CLI_TARGET")
			return 2
		}
		for name := range set {
			if name != "ch" && name != "json" && name != "core" && name != "gui" {
				fmt.Fprintln(os.Stderr, "Chapter 20 does not accept -"+name)
				return 2
			}
		}
	}
	if *chapter == 21 {
		if strings.TrimSpace(*gui) == "" {
			fmt.Fprintln(os.Stderr, "Chapter 21 requires -gui GUI_TARGET plus a positional CLI_TARGET")
			return 2
		}
		for name := range set {
			if name != "ch" && name != "json" && name != "gui" {
				fmt.Fprintln(os.Stderr, "Chapter 21 does not accept -"+name)
				return 2
			}
		}
	}
	if *chapter == 19 {
		for name := range set {
			if name != "ch" && name != "core" && name != "json" {
				fmt.Fprintln(os.Stderr, "Chapter 19 does not use -"+name)
				return 2
			}
		}
		info, err := os.Stat(filepath.Join(*core, "go.mod"))
		if strings.TrimSpace(*core) == "" || err != nil || !info.Mode().IsRegular() {
			fmt.Fprintln(os.Stderr, "Chapter 19 requires -core with an importable module")
			return 2
		}
	}
	// No historical target defaults and no implicit source inference from a binary.
	if (*chapter != 1 && *chapter != 4 && *chapter != 5 && *chapter != 6 && *chapter != 7 && *chapter != 8 && *chapter != 9 && *chapter != 10 && *chapter != 11 && *chapter != 12 && *chapter != 13 && *chapter != 14 && *chapter != 15 && *chapter != 16 && *chapter != 18 && *chapter != 19 && *chapter != 20 && *chapter != 21) || flags.NArg() != 1 || flags.Arg(0) == "" || (*chapter == 1 && (chapter4Flags || chapter5Flags)) || (*chapter == 4 && chapter5Flags) || (*chapter != 6 && *chapter != 7 && *chapter != 8 && *chapter != 9 && *chapter != 10 && *chapter != 11 && *chapter != 12 && *chapter != 13 && chapter6Flags) || (*chapter != 7 && *chapter != 8 && *chapter != 9 && *chapter != 10 && *chapter != 11 && *chapter != 12 && *chapter != 13 && *chapter != 18 && *chapter != 20 && *chapter != 21 && chapter7Flags) {
		fmt.Fprintln(os.Stderr, "usage: grade-edition3 -ch 21 -gui GUI_TARGET [-json] CLI_TARGET; or -ch 20 -core DIR -gui GUI_TARGET [-json] CLI_TARGET; or -ch 19 -core DIR [-json] CLI_TARGET; or -ch 18 -core DIR -gui GUI_TARGET -status-prices I,W,R,O [-status-second-prices I,W,R,O] [-json] CLI_TARGET; or -ch 15 -cache-prices I,W,R,O [-cache-current PATH] [-cache-prior PATH] [-json] GUI_TARGET; or -ch 14|16 [-json] GUI_TARGET; or -ch 1 [-json] CLI_TARGET; or -ch 4|5|6|7|8|9|10|11|12|13 -core DIR -consumer DIR [-consumer-case FILE] [-workflow DIR [-media-models FILE]] [-stream-exercise DIR -stream-model MODEL] [-gui DIR -gui-model MODEL] [-json] target")
		return 2
	}
	var example gradeedition2.ConsumerCase
	if *chapter == 4 || *chapter == 5 || *chapter == 6 || *chapter == 7 || *chapter == 8 || *chapter == 9 || *chapter == 10 || *chapter == 11 || *chapter == 12 || *chapter == 13 {
		dirs := []string{*core, *consumer}
		if *chapter == 5 || *chapter == 6 || *chapter == 7 || *chapter == 8 || *chapter == 9 || *chapter == 10 || *chapter == 11 || *chapter == 12 || *chapter == 13 {
			dirs = append(dirs, *workflow)
		}
		if *chapter == 6 || *chapter == 7 || *chapter == 8 || *chapter == 9 || *chapter == 10 || *chapter == 11 || *chapter == 12 || *chapter == 13 {
			dirs = append(dirs, *streamExercise)
			if strings.TrimSpace(*streamModel) == "" {
				fmt.Fprintln(os.Stderr, "grader: explicit nonblank -stream-model is required for Chapters 6/7/8/9/10/11/12/13")
				return 2
			}
		}
		if *chapter == 7 || *chapter == 8 || *chapter == 9 || *chapter == 10 || *chapter == 11 || *chapter == 12 || *chapter == 13 {
			dirs = append(dirs, *gui)
			if strings.TrimSpace(*guiModel) == "" {
				fmt.Fprintln(os.Stderr, "grader: explicit nonblank -gui-model is required for Chapters 7/8/9/10/11/12/13")
				return 2
			}
		}
		for _, dir := range dirs {
			info, err := os.Stat(filepath.Join(dir, "go.mod"))
			if dir == "" || err != nil || !info.Mode().IsRegular() {
				fmt.Fprintln(os.Stderr, "grader: explicit core and consumer (plus workflow for Chapters 5/6/7/8/9/10/11/12/13, stream-exercise for Chapters 6/7/8/9/10/11/12/13, and gui for Chapters 7/8/9/10/11/12/13) module directories containing go.mod are required")
				return 2
			}
		}
		var err error
		example, err = gradeedition2.ReadConsumerCase(*consumerCase)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			return 2
		}
	}
	var models map[string]string
	if *chapter == 5 || *chapter == 6 || *chapter == 7 || *chapter == 8 || *chapter == 9 || *chapter == 10 || *chapter == 11 || *chapter == 12 || *chapter == 13 {
		var err error
		models, err = gradeedition2.ReadMediaModels(*mediaModels)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			return 2
		}
	}
	bin, cleanup, err := grade.Build(flags.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "grader: %v\n", err)
		return 2
	}
	defer cleanup()
	var report grade.Report
	if *chapter == 4 {
		// Reuse the historical checks; only separate executable and source identities.
		report, err = gradeedition2.Chapter4(bin, *core, *consumer, example)
	} else if *chapter == 5 {
		report, err = gradeedition2.Chapter5(bin, *core, *consumer, *workflow, example, models)
	} else if *chapter == 6 {
		report, err = gradeedition2.Chapter6(bin, *core, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	} else if *chapter == 7 {
		report, err = gradeedition2.Chapter7(bin, *core, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	} else if *chapter == 8 {
		report, err = gradeedition2.Chapter8(bin, *core, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	} else if *chapter == 9 {
		report, err = gradeedition2.Chapter9(bin, *core, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	} else if *chapter == 10 {
		report, err = gradeedition2.Chapter10(bin, *core, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	} else if *chapter == 11 {
		report, err = gradeedition2.Chapter11(bin, *core, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	} else if *chapter == 12 {
		report, err = gradeedition2.Chapter12(bin, *core, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	} else if *chapter == 13 {
		report, err = gradeedition2.Chapter13(bin, *core, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	} else if *chapter == 14 {
		report, err = gradeedition2.Chapter14(bin)
	} else if *chapter == 15 {
		report, err = gradeedition2.Chapter15(bin, *cacheCurrent, *cachePrior, prices)
	} else if *chapter == 16 {
		report, err = gradeedition2.Chapter16(bin)
	} else if *chapter == 20 {
		var guiBin string
		var cleanupGUI func()
		guiBin, cleanupGUI, err = grade.Build(*gui)
		if err == nil {
			defer cleanupGUI()
			report, err = gradeedition2.Chapter20Targets(bin, *core, filepath.Join(*core, "gui"), guiBin)
		}
	} else if *chapter == 21 {
		report, err = gradeedition2.Chapter21CLI(bin, *gui)
	} else if *chapter == 19 {
		report, err = gradeedition2.Chapter19CLI(bin, *core)
	} else if *chapter == 18 {
		var guiBin string
		var cleanupGUI func()
		guiBin, cleanupGUI, err = grade.Build(*gui)
		if err == nil {
			defer cleanupGUI()
			report, err = gradeedition2.Chapter18Targets(bin, guiBin, *core, statusRates, secondRates)
		}
	} else {
		report, err = gradeedition3.Chapter1(bin)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "grader: %v\n", err)
		return 2
	}
	if *asJSON {
		if err := report.WriteJSON(os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			return 2
		}
	} else {
		report.WriteText(os.Stdout)
	}
	if !report.Passed {
		return 1
	}
	return 0
}
