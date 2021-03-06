package cloc

import (
	"os"
	"path/filepath"
	"sync"
)

// Processor represents a process instance
type Processor struct {
	langs *DefinedLanguages
	opts  *Options
	paths []string
}

// Result returns the analysis results
type Result struct {
	Total     *Language
	Files     map[string]*File
	Languages map[string]*Language
}

type syncMap struct {
	m map[string]*File
	sync.RWMutex
}

func newSyncMap(n int) *syncMap {
	return &syncMap{
		m: make(map[string]*File, n),
	}
}

// NewProcessor returns a processor.
func NewProcessor(langs *DefinedLanguages, options *Options, paths []string) *Processor {
	return &Processor{
		langs: langs,
		opts:  options,
		paths: paths,
	}
}

// Analyze starts files analysis.
func (p *Processor) Analyze() (*Result, error) {
	total := NewLanguage("TOTAL", []string{}, [][]string{{"", ""}})

	// List all files and init languages
	// ---------------------------------
	languages, err := p.initLanguages()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	// Analyze of each filen by language
	// ---------------------------------
	syncFiles := newSyncMap(getTotalFiles(languages))
	// files := make(map[string]*File, getTotalFiles(languages))

	for _, language := range languages {
		wg.Add(1)
		go func(language, total *Language, p *Processor, wg *sync.WaitGroup) {
			defer wg.Done()

			for _, file := range language.Files {
				// File analysis
				// -------------
				f := NewFile(file, language.Name)
				f.analyze(language, p.opts)

				// Update language
				// ---------------
				language.Total++
				language.Size += f.Size
				language.Blanks += f.Blanks
				language.Code += f.Code
				language.Comments += f.Comments
				language.Lines += f.Lines

				// Bad performance?
				syncFiles.Lock()
				syncFiles.m[file] = f
				syncFiles.Unlock()
			}

			// Totals
			// ------
			nbFiles := int32(len(language.Files))
			if len(language.Files) <= 0 {
				return
			}
			total.Size += language.Size
			total.Total += nbFiles
			total.Blanks += language.Blanks
			total.Comments += language.Comments
			total.Code += language.Code
			total.Lines += language.Lines
		}(language, total, p, &wg)
	}

	wg.Wait()

	return &Result{
		Total:     total,
		Files:     syncFiles.m,
		Languages: languages,
	}, nil
}

// initLanguages lists all files form paths and inits languages.
func (p *Processor) initLanguages() (result map[string]*Language, err error) {
	result = make(map[string]*Language)
	filesCache := make(map[string]struct{})

	for _, root := range p.paths {
		vcsInRoot := isVCSDir(root)
		err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			// Check if the language is analysable
			// -----------------------------------
			if ok := isLanguageAnalysable(path, vcsInRoot, p.opts); !ok {
				return nil
			}

			// Check file extension
			// --------------------
			if ext, ok := getExtension(path, p.opts); ok {
				// Get Language
				// ------------
				if lang, ok := Extensions[ext]; ok {
					// Check Options
					// -------------
					if ok := checkFileOptions(path, lang, p.opts, filesCache); ok {
						// Add to languages list
						// ---------------------
						if _, ok := result[lang]; !ok {
							result[lang] = NewLanguage(
								p.langs.Langs[lang].Name,
								p.langs.Langs[lang].lineComments,
								p.langs.Langs[lang].multiLines)
						}
						result[lang].Files = append(result[lang].Files, path)
					}
				}
			}

			return nil
		})
	}

	return result, err
}
