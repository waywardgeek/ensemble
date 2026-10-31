// Command grade-edition2 runs the separately maintained second-edition grader.
// Invoke it from the repository root: grade-edition2 -ch 1 [submission-directory].
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/waywardgeek/ensemble/internal/grade"
	"github.com/waywardgeek/ensemble/internal/gradeedition2"
)

func main() {
	os.Exit(run())
}

// Return an exit status so deferred build cleanup runs even after a failed grade.
func run() int {
	chapter := flag.Int("ch", 1, "Edition 2 chapter to grade")
	asJSON := flag.Bool("json", false, "emit the report as JSON")
	consumer := flag.String("consumer", "", "Chapter 4–13 separate consumer module (default: ch05 under or beside submission)")
	consumerCase := flag.String("consumer-case", "", "Chapter 4–13 JSON case file (default: accepted shout example)")
	workflow := flag.String("workflow", "", "Chapter 5–13 workflow module (default: ch06 under or beside submission)")
	mediaModels := flag.String("media-models", "", "Chapter 5–13 JSON vendor/model map (default: historical accepted fixtures)")
	streamModel := flag.String("stream-model", "claude-sonnet-5", "Chapter 6–13 declared Anthropic streaming fixture (historical default)")
	streamExercise := flag.String("stream-exercise", "", "Chapter 6–13 external shape module (default: ch07 under or beside submission)")
	gui := flag.String("gui", "", "Chapter 7–20 GUI module (default: gui under submission)")
	guiModel := flag.String("gui-model", "claude-sonnet-5", "Chapter 7–13 declared Anthropic GUI fixture (historical default)")
	recallData := flag.String("recall-data", "cr", "Chapter 16 grader corpus root relative to scratch cwd (legacy: .)")
	cacheCurrent := flag.String("cache-current", "cr/cache-current.txt", "Chapter 17 current capture relative to scratch cwd")
	cachePrior := flag.String("cache-prior", "cr/cache-prior.txt", "Chapter 17 prior capture relative to scratch cwd")
	cachePrices := flag.String("cache-prices", "2,2.5,0.2,10", "Chapter 17 Sonnet-5 dollars/million: fresh,write,read,output (accepted October 11 table)")
	statusPrices := flag.String("status-prices", "2,2.5,0.2,10", "Chapter 20 Sonnet-5 dollars/million: fresh,write,read,output")
	statusSecondPrices := flag.String("status-second-prices", "unknown", "Chapter 20 Opus 4.6 rates in same order, or unknown (frozen accepted table)")
	sandboxLibrary := flag.String("sandbox-library", "ensemble", "Chapter 21 public package relative to submitted module (legacy: .)")
	flag.Parse()
	if *chapter < 1 || *chapter > 21 || *chapter == 19 || flag.NArg() > 1 || (*chapter < 4 && (*consumer != "" || *consumerCase != "")) || (*chapter < 5 && (*workflow != "" || *mediaModels != "")) || (*chapter < 6 && *streamExercise != "") || (*chapter < 7 && *gui != "") {
		fmt.Fprintln(os.Stderr, "usage: grade-edition2 -ch 1..18|20..21 [-json] [-consumer DIR -consumer-case FILE] [-workflow DIR -media-models FILE] [-stream-exercise DIR -stream-model MODEL] [-gui DIR -gui-model MODEL] [submission-directory]; Chapter 19 is ungraded")
		return 2
	}
	path := flag.Arg(0)
	if path == "" {
		path = fmt.Sprintf("./solutions/edition-2/ch%02d", *chapter)
	}
	var example gradeedition2.ConsumerCase
	if *chapter >= 4 && *chapter < 14 {
		var err error
		example, err = gradeedition2.ReadConsumerCase(*consumerCase)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			return 2
		}
		if *consumer == "" {
			*consumer = appDirectory(path, "ch05")
		}
	}
	var models map[string]string
	if *chapter >= 5 && *chapter < 14 {
		var err error
		models, err = gradeedition2.ReadMediaModels(*mediaModels)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grader: %v\n", err)
			return 2
		}
		if *workflow == "" {
			*workflow = appDirectory(path, "ch06")
		}
	}
	if *chapter >= 6 && *chapter < 14 && *streamExercise == "" {
		*streamExercise = appDirectory(path, "ch07")
	}
	if *chapter >= 7 && *gui == "" {
		*gui = filepath.Join(path, "gui")
	}
	buildPath := path
	if *chapter >= 14 && *chapter != 21 {
		buildPath = *gui
	}
	bin, cleanup, err := grade.Build(buildPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grader: %v\n", err)
		return 2
	}
	defer cleanup()
	var report grade.Report
	switch *chapter {
	case 21:
		report, err = gradeedition2.Chapter21(bin, path, *sandboxLibrary)
	case 20:
		var rates [4]float64
		var second *[4]float64
		rates, err = gradeedition2.ParseCachePrices(*statusPrices)
		if err == nil && *statusSecondPrices != "unknown" {
			var value [4]float64
			value, err = gradeedition2.ParseCachePrices(*statusSecondPrices)
			second = &value
		}
		if err == nil {
			report, err = gradeedition2.Chapter20(bin, path, rates, second)
		}
	case 1:
		var res *grade.RunResult
		res, err = grade.Run(bin)
		if err == nil {
			report = gradeedition2.Chapter1(res)
		}
	case 2:
		report, err = gradeedition2.Chapter2(bin)
	case 3:
		report, err = gradeedition2.Chapter3(bin)
	case 4:
		report, err = gradeedition2.Chapter4(bin, path, *consumer, example)
	case 5:
		report, err = gradeedition2.Chapter5(bin, path, *consumer, *workflow, example, models)
	case 6:
		report, err = gradeedition2.Chapter6(bin, path, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	case 7:
		report, err = gradeedition2.Chapter7(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	case 8:
		report, err = gradeedition2.Chapter8(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	case 9:
		report, err = gradeedition2.Chapter9(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	case 10:
		report, err = gradeedition2.Chapter10(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	case 11:
		report, err = gradeedition2.Chapter11(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	case 12:
		report, err = gradeedition2.Chapter12(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
	case 18:
		report, err = gradeedition2.Chapter18(bin)
	case 17:
		var rates [4]float64
		rates, err = gradeedition2.ParseCachePrices(*cachePrices)
		if err == nil {
			report, err = gradeedition2.Chapter17(bin, *cacheCurrent, *cachePrior, rates)
		}
	case 16:
		report, err = gradeedition2.Chapter16(bin, path, *gui, *recallData)
	case 15:
		report, err = gradeedition2.Chapter15(bin)
	case 14:
		report, err = gradeedition2.Chapter14(bin)
	case 13:
		report, err = gradeedition2.Chapter13(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
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

// Historical exports place applications either inside the core tree or beside
// it. Discover only those established layouts; an absent app still fails grading.
func appDirectory(core, name string) string {
	for _, dir := range []string{filepath.Join(core, name), filepath.Join(core, "..", name)} {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
	}
	return filepath.Join(core, name)
}
