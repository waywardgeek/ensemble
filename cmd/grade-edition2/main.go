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
	executableOnly := flag.Bool("executable-only", false, "Chapters 5–16, 18–23, 26 executable checks only; Chapter 26 grades only server behavior using an independent peer")
	jobsOnly := flag.Bool("jobs-only", false, "Chapter 4 executable jobs only; excludes the separate Go consumer checks")
	asJSON := flag.Bool("json", false, "emit the report as JSON")
	consumer := flag.String("consumer", "", "Chapter 4–13 separate consumer module (default: ch05 under or beside submission)")
	consumerCase := flag.String("consumer-case", "", "Chapter 4–13 JSON case file (default: accepted shout example)")
	workflow := flag.String("workflow", "", "Chapter 5–13 workflow module (default: ch06 under or beside submission)")
	mediaModels := flag.String("media-models", "", "Chapter 5–13 JSON vendor/model map (default: historical accepted fixtures)")
	streamModel := flag.String("stream-model", "claude-sonnet-5", "Chapter 6–13 declared Anthropic streaming fixture (historical default)")
	streamExercise := flag.String("stream-exercise", "", "Chapter 6–13 external shape module (default: ch07 under or beside submission)")
	gui := flag.String("gui", "", "Chapter 7–23 GUI module (default: gui under submission)")
	guiModel := flag.String("gui-model", "claude-sonnet-5", "Chapter 7–13 declared Anthropic GUI fixture (historical default)")
	recallData := flag.String("recall-data", "cr", "Chapter 23 grader corpus root relative to scratch cwd (legacy: .)")
	cacheCurrent := flag.String("cache-current", "cr/cache-current.txt", "Chapter 15 current capture relative to scratch cwd")
	cachePrior := flag.String("cache-prior", "cr/cache-prior.txt", "Chapter 15 prior capture relative to scratch cwd")
	cachePrices := flag.String("cache-prices", "2,2.5,0.2,10", "Chapter 15 Sonnet-5 dollars/million: fresh,write,read,output (accepted October 11 table)")
	fleetSkills := flag.String("fleet-skills", "", "Chapters 20/21 executable scope: absolute shipped skill directory")
	statusSkills := flag.String("status-skills", "", "Chapter 18 executable scope: shipped skill directory")
	statusPrices := flag.String("status-prices", "2,2.5,0.2,10", "Chapter 18 Sonnet-5 dollars/million: fresh,write,read,output")
	statusSecondPrices := flag.String("status-second-prices", "unknown", "Chapter 18 Opus 4.6 rates in same order, or unknown (frozen accepted table)")
	sandboxLibrary := flag.String("sandbox-library", "ensemble", "Chapter 19 public package relative to submitted module (legacy: .)")
	reuseLegacy := flag.Bool("reuse-legacy", false, "Chapter 20: use the historical same-module public construction API")
	gatewayPeer := flag.String("gateway-peer", "", "Chapter 26 executable scope: absolute independent Go peer source directory (not the submitted library)")
	gatewayWelcome := flag.String("gateway-welcome", "QueueDepth", "Chapter 26 public welcome queue field (legacy: QueuedCount)")
	flag.Parse()
	if *chapter < 1 || (*chapter > 23 && *chapter != 26) || *chapter == 17 || flag.NArg() > 1 || (*chapter < 4 && (*consumer != "" || *consumerCase != "")) || (*chapter < 5 && (*workflow != "" || *mediaModels != "")) || (*chapter < 6 && *streamExercise != "") || (*chapter < 7 && *gui != "") {
		fmt.Fprintln(os.Stderr, "usage: grade-edition2 -ch 1..16|18..23|26 [-json] [-consumer DIR -consumer-case FILE] [-workflow DIR -media-models FILE] [-stream-exercise DIR -stream-model MODEL] [-gui DIR -gui-model MODEL] [submission-directory]; Chapters 17, 24 and 25 are ungraded")
		return 2
	}
	path := flag.Arg(0)
	if path == "" {
		path = fmt.Sprintf("./solutions/ch%02d", *chapter)
	}
	if *jobsOnly && (*chapter != 4 || *consumer != "" || *consumerCase != "") {
		fmt.Fprintln(os.Stderr, "-jobs-only requires Chapter 4 and excludes consumer flags")
		return 2
	}
	if *executableOnly && ((*chapter != 5 && *chapter != 6 && *chapter != 7 && *chapter != 8 && *chapter != 9 && *chapter != 10 && *chapter != 11 && *chapter != 12 && *chapter != 13 && *chapter != 14 && *chapter != 15 && *chapter != 16 && *chapter != 18 && *chapter != 19 && *chapter != 20 && *chapter != 21 && *chapter != 22 && *chapter != 23 && *chapter != 26) || *jobsOnly || *consumer != "" || *consumerCase != "" || *workflow != "" || (*chapter == 6 && *streamExercise == "") || ((*chapter == 7 || *chapter == 8 || *chapter == 9 || *chapter == 14 || *chapter == 15 || *chapter == 16 || *chapter == 18) && *gui == "")) {
		fmt.Fprintln(os.Stderr, "-executable-only requires Chapter 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 18, 19, 20, 21, 22, 23 or 26, excludes jobs-only/consumer/workflow flags; Chapter 6 requires -stream-exercise EXECUTABLE; Chapters 7/8/9 require -gui STATIC-DIRECTORY; Chapters 14/15/16/18 require -gui EXECUTABLE; Chapter 18 also requires -status-skills DIRECTORY")
		return 2
	}
	if (*gatewayPeer != "" && (!*executableOnly || *chapter != 26)) || (*executableOnly && *chapter == 26 && (!filepath.IsAbs(*gatewayPeer) || *gui != "" || *fleetSkills != "" || *statusSkills != "" || *reuseLegacy || *streamExercise != "" || *sandboxLibrary != "ensemble" || *gatewayWelcome != "QueueDepth" || *recallData != "cr" || *mediaModels != "")) {
		fmt.Fprintln(os.Stderr, "Chapter 26 -executable-only requires absolute -gateway-peer DIRECTORY and excludes GUI, fleet, provider, Go library and historical welcome flags")
		return 2
	}
	if *executableOnly && *chapter == 22 && (*gui != "" || *fleetSkills != "" || *reuseLegacy || *streamExercise != "" || *statusSkills != "" || *sandboxLibrary != "ensemble") {
		fmt.Fprintln(os.Stderr, "Chapter 22 -executable-only excludes GUI, fleet, streaming and Go library flags")
		return 2
	}
	if *executableOnly && *chapter == 23 && (*gui != "" || *fleetSkills != "" || *reuseLegacy || *streamExercise != "" || *statusSkills != "" || *sandboxLibrary != "ensemble" || *recallData != "cr") {
		fmt.Fprintln(os.Stderr, "Chapter 23 -executable-only excludes GUI, fleet, streaming, Go library and recall-data flags")
		return 2
	}
	if *executableOnly && *chapter == 19 && (*gui != "" || *sandboxLibrary != "ensemble") {
		fmt.Fprintln(os.Stderr, "Chapter 19 -executable-only excludes GUI and Go sandbox-library flags")
		return 2
	}
	if (*fleetSkills != "" && (!*executableOnly || (*chapter != 20 && *chapter != 21))) || (*executableOnly && (*chapter == 20 || *chapter == 21) && (!filepath.IsAbs(*fleetSkills) || *gui != "" || *reuseLegacy)) {
		fmt.Fprintln(os.Stderr, "grader: Chapters 20/21 executable scope requires an absolute -fleet-skills and excludes GUI/Go consumer flags")
		return 2
	}
	if (*statusSkills != "" && (!*executableOnly || *chapter != 18)) || (*executableOnly && *chapter == 18 && *statusSkills == "") {
		fmt.Fprintln(os.Stderr, "-status-skills is required only for Chapter 18 -executable-only")
		return 2
	}
	var example gradeedition2.ConsumerCase
	if *chapter >= 4 && *chapter < 14 && !*jobsOnly && !*executableOnly {
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
	if *chapter == 13 && *executableOnly {
		buildPath = filepath.Join(path, "scripts", "tts-harness.sh")
	}
	switch *chapter {
	case 14, 15, 16, 18, 22, 23:
		buildPath = *gui
	}
	if (*chapter == 18 || *chapter == 22 || *chapter == 23) && *executableOnly {
		buildPath = path
	}
	if *chapter == 26 && !*executableOnly {
		buildPath = filepath.Join(path, "cmd", "gateway")
	}
	bin, cleanup, err := grade.Build(buildPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grader: %v\n", err)
		return 2
	}
	defer cleanup()
	var report grade.Report
	switch *chapter {
	case 26:
		if *executableOnly {
			report, err = gradeedition2.Chapter26Executable(bin, *gatewayPeer)
		} else {
			report, err = gradeedition2.Chapter26(bin, path, *gatewayWelcome)
		}
	case 21:
		if *executableOnly {
			report, err = gradeedition2.Chapter21Executable(bin, *fleetSkills)
		} else {
			report, err = gradeedition2.Chapter21(bin, *gui)
		}
	case 20:
		if *executableOnly {
			report, err = gradeedition2.Chapter20Executable(bin, *fleetSkills)
		} else {
			report, err = gradeedition2.Chapter20(bin, path, *gui, *reuseLegacy)
		}
	case 19:
		if *executableOnly {
			report, err = gradeedition2.Chapter19Executable(bin)
		} else {
			report, err = gradeedition2.Chapter19(bin, path, *sandboxLibrary)
		}
	case 18:
		if *executableOnly {
			var guiBin string
			var cleanGUI func()
			guiBin, cleanGUI, err = grade.Build(*gui)
			if err == nil {
				defer cleanGUI()
				report, err = gradeedition2.Chapter18Executable(bin, guiBin, *statusSkills)
			}
			break
		}
		var rates [4]float64
		var second *[4]float64
		rates, err = gradeedition2.ParseCachePrices(*statusPrices)
		if err == nil && *statusSecondPrices != "unknown" {
			var value [4]float64
			value, err = gradeedition2.ParseCachePrices(*statusSecondPrices)
			second = &value
		}
		if err == nil {
			report, err = gradeedition2.Chapter18(bin, path, rates, second)
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
		if *jobsOnly {
			report, err = gradeedition2.Chapter4Jobs(bin)
		} else {
			report, err = gradeedition2.Chapter4(bin, path, *consumer, example)
		}
	case 5:
		if *executableOnly {
			report = gradeedition2.Chapter5Executable(bin, models)
		} else {
			report, err = gradeedition2.Chapter5(bin, path, *consumer, *workflow, example, models)
		}
	case 6:
		if *executableOnly {
			report = gradeedition2.Chapter6Executable(bin, *streamExercise, *streamModel)
		} else {
			report, err = gradeedition2.Chapter6(bin, path, *consumer, *workflow, *streamExercise, *streamModel, example, models)
		}
	case 7:
		if *executableOnly {
			report = gradeedition2.Chapter7Executable(bin, *gui, *guiModel)
		} else {
			report, err = gradeedition2.Chapter7(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
		}
	case 8:
		if *executableOnly {
			report = gradeedition2.Chapter8Executable(bin, *gui, *guiModel)
		} else {
			report, err = gradeedition2.Chapter8(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
		}
	case 9:
		if *executableOnly {
			report = gradeedition2.Chapter9Executable(bin, *gui, *guiModel)
		} else {
			report, err = gradeedition2.Chapter9(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
		}
	case 10:
		if *executableOnly {
			report = gradeedition2.Chapter10Executable(bin)
		} else {
			report, err = gradeedition2.Chapter10(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
		}
	case 11:
		if *executableOnly {
			report = gradeedition2.Chapter11Executable(bin)
		} else {
			report, err = gradeedition2.Chapter11(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
		}
	case 12:
		if *executableOnly {
			report = gradeedition2.Chapter12Executable(bin)
		} else {
			report, err = gradeedition2.Chapter12(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
		}
	case 16:
		if *executableOnly {
			report, err = gradeedition2.Chapter16Executable(bin)
		} else {
			report, err = gradeedition2.Chapter16(bin)
		}
	case 15:
		var rates [4]float64
		rates, err = gradeedition2.ParseCachePrices(*cachePrices)
		if err == nil {
			if *executableOnly {
				report, err = gradeedition2.Chapter15Executable(bin, rates)
			} else {
				report, err = gradeedition2.Chapter15(bin, *cacheCurrent, *cachePrior, rates)
			}
		}
	case 23:
		if *executableOnly {
			report, err = gradeedition2.Chapter23Executable(bin)
		} else {
			report, err = gradeedition2.Chapter23(bin, path, *gui, *recallData)
		}
	case 22:
		if *executableOnly {
			report, err = gradeedition2.Chapter22Executable(bin)
		} else {
			report, err = gradeedition2.Chapter22(bin)
		}
	case 14:
		if *executableOnly {
			report, err = gradeedition2.Chapter14Executable(bin)
		} else {
			report, err = gradeedition2.Chapter14(bin)
		}
	case 13:
		if *executableOnly {
			report = gradeedition2.Chapter13Executable(path)
		} else {
			report, err = gradeedition2.Chapter13(bin, path, *gui, *guiModel, *consumer, *workflow, *streamExercise, *streamModel, example, models)
		}
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
