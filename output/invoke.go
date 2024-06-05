package output

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/gofoji/plates"
	"github.com/gofoji/plates/plush"
	"github.com/rs/zerolog"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/foji"
	"github.com/gofoji/foji/runtime"
	"github.com/gofoji/foji/stringlist"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrMissingRequirement = Error("requires")
	ErrNotNeeded          = Error("not needed")
	ErrPermExists         = Error("file exists")
)

type FuncMapper interface {
	Funcs() plates.FuncMap
}

type Initializer interface {
	Init() error
}

type TemplateInitializer interface {
	Init(*plates.Factory) error
}

const PermPrefix = "!"

func (p ProcessRunner) process(tm stringlist.StringMap, data any) error {
	f, ok := data.(FuncMapper)
	if ok {
		p.AddFuncs(f.Funcs())
	}

	for targetFile, templateFile := range tm {
		err := p.doInit(data)
		if err != nil {
			return err
		}

		err = p.template(targetFile, templateFile, data)
		if err != nil {
			if !errors.Is(err, ErrPermExists) {
				return err
			}

			p.l.Warn().Str("target", targetFile).Str("template", templateFile).Msg("skipped, output file exists")
		}
	}

	return nil
}

func (p ProcessRunner) doInit(data any) error {
	init, ok := data.(Initializer)
	if ok {
		err := init.Init()
		if err != nil {
			return fmt.Errorf("error initializing context: %w", err)
		}
	}

	tInit, ok := data.(TemplateInitializer)
	if ok {
		err := tInit.Init(p.Factory)
		if err != nil {
			return fmt.Errorf("error initializing context: %w", err)
		}
	}

	return nil
}

func checkPermanentFlag(outputFile string) (bool, string) {
	if strings.HasPrefix(outputFile, PermPrefix) {
		return true, outputFile[1:]
	}

	return false, outputFile
}

func templateEngine() *plates.Factory {
	return plates.New("foji").
		AddMatcherFunc(plates.MatchText, plates.MatchHTML, plush.Match).
		AddFuncs(runtime.Funcs, sprig.GenericFuncMap()).
		DefaultFunc(plates.TextParser)
}

type ProcessRunner struct {
	l        zerolog.Logger
	dir      string
	fn       cfg.FileHandler
	simulate bool
	*plates.Factory
}

func NewProcessRunner(dir string, fn cfg.FileHandler, l zerolog.Logger, simulate bool) ProcessRunner {
	if dir != "" {
		if !strings.HasSuffix(dir, string(os.PathSeparator)) {
			dir += string(os.PathSeparator)
		}
	}

	p := ProcessRunner{
		Factory:  templateEngine(),
		l:        l,
		simulate: simulate,
		fn:       fn,
		dir:      dir,
	}
	p.FileReaderFunc(p.loadLocalOrEmbed)

	return p
}

func (p ProcessRunner) template(outputFile, templateFile string, data any) error {
	permFile, outputFile := checkPermanentFlag(outputFile)

	outputFile, err := p.From(outputFile).To(data)
	if err != nil {
		return fmt.Errorf("mapping output filename:%w", err)
	}

	l := p.l.With().Str("target", outputFile).Str("template", templateFile).Logger()
	l.Info().Msg("executing template")

	outputFile = p.dir + outputFile
	if permFile && fileExists(outputFile) {
		return ErrPermExists
	}

	if p.simulate {
		l.Info().Msg("simulated")

		return nil
	}

	// Provides the current template file to the context
	p.AddFuncs(map[string]any{"templateFile": func() string { return templateFile }})

	err = p.FromFile(templateFile).ToFile(outputFile, data)
	if err != nil {
		if errors.Is(err, ErrNotNeeded) {
			l.Info().Err(err).Msg("skipped")

			return nil
		}

		if errors.Is(err, ErrMissingRequirement) {
			return err //nolint:wrapcheck
		}

		return fmt.Errorf("executing template: %q: %w", templateFile, err)
	}

	l.Debug().Msg("wrote output")

	return p.postProcess(outputFile)
}

func (p ProcessRunner) postProcess(outputFile string) error {
	if p.fn != nil {
		err := p.fn(outputFile)
		if err != nil {
			return fmt.Errorf("post processing file: %s: %w", outputFile, err)
		}
	}

	return nil
}

func (p ProcessRunner) loadLocalOrEmbed(filename string) ([]byte, error) {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			p.l.Debug().Str("filename", filename).Msg("loading from embed")

			return foji.Default(filename)
		}

		return nil, fmt.Errorf("error accessing file: %s: %w", filename, err)
	}

	return ioutil.ReadFile(filename)
}

func hasAnyOutput(o cfg.Output, outputs ...string) bool {
	for _, k := range outputs {
		if len(o[k]) > 0 {
			return true
		}
	}

	return false
}

func fileExists(filename string) bool {
	fileInfo, err := os.Stat(filename)

	return err == nil && fileInfo.Mode().IsRegular()
}
