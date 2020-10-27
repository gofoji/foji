package output

import (
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/runtime"
	"github.com/gofoji/foji/stringlist"
	"github.com/gofoji/foji/tpl"
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
	Funcs() template.FuncMap
}

type Initializer interface {
	Init() error
}

func fileExists(filename string) bool {
	fileInfo, err := os.Stat(filename)
	return err == nil && fileInfo.Mode().IsRegular()
}

const PermPrefix = "!"

func invokeProcess(tm stringlist.StringMap, dir string, fn cfg.FileHandler, logger logrus.FieldLogger, data interface{}, simulate bool) error {
	for targetFile, templateFile := range tm {
		err := invokeTemplate(logger, dir, targetFile, templateFile, data, fn, simulate)
		if err != nil {
			if !errors.Is(err, ErrPermExists) {
				return err
			}
			logger.WithField("target", targetFile).WithField("template", templateFile).Warn("skipped, output file exists")
		}
	}
	return nil
}

func invokeTemplate(logger logrus.FieldLogger, dir, targetFile, templateFile string, data interface{}, fn cfg.FileHandler, simulate bool) error {
	logger.WithField("target", targetFile).WithField("template", templateFile).Info("executing template")

	init, ok := data.(Initializer)
	if ok {
		err := init.Init()
		if err != nil {
			return err
		}
	}

	permFile := false
	if strings.HasPrefix(targetFile, PermPrefix) {
		targetFile = targetFile[1:]
		permFile = true
	}

	if dir != "" {
		if !strings.HasSuffix(dir, string(os.PathSeparator)) {
			dir = dir + string(os.PathSeparator)
		}
		targetFile = dir + targetFile
	}

	targetFile, err := tpl.New("outputMapper").Funcs(runtime.Funcs).Funcs(sprig.TxtFuncMap()).From(targetFile).To(data)
	if err != nil {
		return errors.Wrap(err, "mapping output filename")
	}

	if permFile && fileExists(targetFile) {
		return ErrPermExists
	}

	if simulate {
		logger.WithField("target", targetFile).WithField("template", templateFile).Info("simulated")
		return nil
	}

	t := tpl.New(templateFile).Funcs(runtime.Funcs)

	f, ok := data.(FuncMapper)
	if ok {
		t.Funcs(f.Funcs())
	}
	tFunc := map[string]interface{}{"templateFile": func() string { return templateFile }}
	t.Funcs(tFunc)
	t.Funcs(sprig.TxtFuncMap())
	err = t.FromFile(templateFile).ToFile(targetFile, data)
	if err != nil {
		if errors.Is(err, ErrNotNeeded) {
			logger.WithField("target", targetFile).WithField("template", templateFile).Info("skipped, " + err.Error())
			return nil
		}

		if errors.Is(err, ErrMissingRequirement) {
			return err
		}

		return errors.Wrap(err, "executing template")
	}

	logger.WithField("target", targetFile).WithField("template", templateFile).Debug("wrote output")

	if fn != nil {
		err = fn(targetFile)
		if err != nil {
			return errors.Wrap(err, "error processing file")
		}
	}
	return nil
}

func hasAnyOutput(o cfg.Output, outputs ...string) bool {
	for _, k := range outputs {
		if len(o[k]) > 0 {
			return true
		}
	}
	return false
}
