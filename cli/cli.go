package cli

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/fabienbellanger/goCodeAnalyser/cloc"
	"github.com/fabienbellanger/goCodeAnalyser/output"
	"github.com/fabienbellanger/goutils"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

// CmdOptions lists all command options.
type CmdOptions struct {
	ByFile         bool
	Debug          bool
	SkipDuplicated bool
	OutputType     string
	ExcludeExt     string
	IncludeLang    string
	MatchDir       string
	NotMatchDir    string
	Sort           string
}

const (
	appName = "Go Code Analyser"
	version = "0.1.0"
)

var (
	// color enables colors in console.
	color aurora.Aurora = aurora.NewAurora(true)

	// cmdOpts stores command options.
	cmdOpts = CmdOptions{}

	rootCommand = &cobra.Command{
		Use:     "goCodeAnalyser [paths]",
		Short:   "goCodeAnalyser [paths]",
		Long:    "goCodeAnalyser [paths]",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			tStart := time.Now()

			// Manage paths
			// ------------
			if len(args) == 0 {
				if err := cmd.Usage(); err != nil {
					goutils.CheckError(err, 1)
				}
				return
			}

			// List of all available languages
			// -------------------------------
			languages := cloc.NewDefinedLanguages()

			// Fill application options
			// ------------------------
			appOpts := fillOptions(cmdOpts, languages)

			// Launch process
			// --------------
			// TODO: To implement
			processor := cloc.NewProcessor(languages, appOpts, args)
			result, err := processor.Analyze()
			if err != nil {
				goutils.CheckError(err, 1)
			}

			// Display results
			// ---------------
			// TODO: Switch output!
			var w output.Writer
			w = output.NewConsole()
			w.Write(result, appOpts)

			fmt.Printf("\nNumber of CPU: %d\n", runtime.NumCPU())
			displayDuration(time.Since(tStart))
		},
	}
)

// Execute starts Cobra.
func Execute() error {
	// Version
	// -------
	rootCommand.SetVersionTemplate(appName + " version " + version + "\n")

	// Flags
	// -----
	rootCommand.Flags().BoolVar(&cmdOpts.ByFile, "files", false, "Display by file")
	rootCommand.Flags().BoolVar(&cmdOpts.Debug, "debug", false, "Display debug log")
	rootCommand.Flags().BoolVar(&cmdOpts.SkipDuplicated, "skip-duplicated", false, "Skip duplicated files")
	rootCommand.Flags().StringVar(&cmdOpts.OutputType, "output-type", "", "Output type [values: default,json,html]")
	rootCommand.Flags().StringVar(&cmdOpts.ExcludeExt, "exclude-ext", "", "Exclude file name extensions (separated commas)")
	rootCommand.Flags().StringVar(&cmdOpts.IncludeLang, "include-lang", "", "Include language name (separated commas)")
	rootCommand.Flags().StringVar(&cmdOpts.MatchDir, "match-dir", "", "Include dir name (regex)")
	rootCommand.Flags().StringVar(&cmdOpts.NotMatchDir, "not-match-dir", "", "Exclude dir name (regex)")
	rootCommand.Flags().StringVar(&cmdOpts.Sort, "sort", "code", "Sort languages based on column [possible values: files, lines, blanks, code, comments or size]")

	// Launch root command
	// -------------------
	if err := rootCommand.Execute(); err != nil {
		return err
	}
	return nil
}

// fillOptions fills applications options from command options.
// TODO: Test
func fillOptions(cmdOpts CmdOptions, languages *cloc.DefinedLanguages) *cloc.Options {
	// Checks sort values
	// ------------------
	if !cloc.CheckSort(cmdOpts.Sort) {
		cmdOpts.Sort = "code"
	}

	opts := cloc.NewOptions()
	opts.ByFile = cmdOpts.ByFile
	opts.Debug = cmdOpts.Debug
	opts.SkipDuplicated = cmdOpts.SkipDuplicated
	opts.Sort = cmdOpts.Sort

	// Excluded extensions
	// -------------------
	for _, ext := range strings.Split(cmdOpts.ExcludeExt, ",") {
		e, ok := cloc.Extensions[ext]
		if ok {
			opts.ExcludeExts[e] = struct{}{}
		}
	}

	// Match or not directory
	// ----------------------
	if cmdOpts.NotMatchDir != "" {
		opts.NotMatchDir = regexp.MustCompile(cmdOpts.NotMatchDir)
	}
	if cmdOpts.MatchDir != "" {
		opts.MatchDir = regexp.MustCompile(cmdOpts.MatchDir)
	}

	// Included languages
	// ------------------
	for _, lang := range strings.Split(cmdOpts.IncludeLang, ",") {
		if _, ok := languages.Langs[lang]; ok {
			opts.IncludeLangs[lang] = struct{}{}
		}
	}

	return opts
}

// displayDuration displays commands execution duration.
func displayDuration(d time.Duration) {
	fmt.Println(color.Sprintf(color.Italic("\nCommand execution time: %v\n"), d))
}
