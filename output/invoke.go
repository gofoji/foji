package output

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/embed"
	"github.com/gofoji/foji/runtime"
	"github.com/gofoji/foji/stringlist"
	"github.com/gofoji/plates"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

const PermPrefix = "!"

func (p ProcessRunner) process(tm stringlist.StringMap, data interface{}) error {
	err := doInit(data)
	if err != nil {
		return err
	}

	f, ok := data.(FuncMapper)
	if ok {
		p.AddFuncs(f.Funcs())
	}

	for targetFile, templateFile := range tm {
		err := p.template(targetFile, templateFile, data)
		if err != nil {
			if !errors.Is(err, ErrPermExists) {
				return err
			}

			p.l.WithField("target", targetFile).WithField("template", templateFile).Warn("skipped, output file exists")
		}
	}

	return nil
}

func doInit(data interface{}) error {
	init, ok := data.(Initializer)
	if ok {
		err := init.Init()
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

func templateEngine() *plates.Wrapper {
	return plates.New("foji").
		AddFuncs(runtime.Funcs, sprig.GenericFuncMap()).
		DefaultFunc(plates.TextParser)
}

type ProcessRunner struct {
	l        logrus.FieldLogger
	dir      string
	fn       cfg.FileHandler
	simulate bool
	*plates.Wrapper
}

func NewProcessRunner(dir string, fn cfg.FileHandler, l logrus.FieldLogger, simulate bool) ProcessRunner {
	if dir != "" {
		if !strings.HasSuffix(dir, string(os.PathSeparator)) {
			dir += string(os.PathSeparator)
		}
	}

	p := ProcessRunner{
		Wrapper:  templateEngine(),
		l:        l,
		simulate: simulate,
		fn:       fn,
		dir:      dir,
	}
	p.FileReaderFunc(p.loadLocalOrEmbed)

	return p
}

func (p ProcessRunner) template(outputFile, templateFile string, data interface{}) error {
	p.l.WithField("target", outputFile).WithField("template", templateFile).Info("executing template")

	permFile, outputFile := checkPermanentFlag(outputFile)

	outputFile, err := p.From(outputFile).To(data)
	if err != nil {
		return fmt.Errorf("mapping output filename:%w", err)
	}

	outputFile = p.dir + outputFile
	if permFile && fileExists(outputFile) {
		return ErrPermExists
	}

	if p.simulate {
		p.l.WithField("target", outputFile).WithField("template", templateFile).Info("simulated")

		return nil
	}

	// Provides the current template file to the context
	p.AddFuncs(map[string]interface{}{"templateFile": func() string { return templateFile }})

	err = p.FromFile(templateFile).ToFile(outputFile, data)
	if err != nil {
		if errors.Is(err, ErrNotNeeded) {
			p.l.WithField("target", outputFile).WithField("template", templateFile).Info("skipped, " + err.Error())

			return nil
		}

		if errors.Is(err, ErrMissingRequirement) {
			return err //nolint:wrapcheck
		}

		return fmt.Errorf("executing template: %w", err)
	}

	p.l.WithField("target", outputFile).WithField("template", templateFile).Debug("wrote output")

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
			p.l.WithField("filename", filename).Debug("loading from embed")

			return embed.Get(filename)
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
