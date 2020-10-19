package tpl

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/gofoji/foji/embed"
	"github.com/pkg/errors"
)

type Wrapper struct {
	t   *template.Template
	err error
}

func loadLocalOrEmbed(filename string) (string, error) {
	_, err := os.Stat(filename)

	if err != nil {
		if os.IsNotExist(err) {
			return embed.Get(filename)
		}

		return "", err
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func New(name string) *Wrapper {
	return &Wrapper{t: template.New(name)}
}

func (t Wrapper) Err() error {
	return t.err
}

func (t *Wrapper) Funcs(funcs ...template.FuncMap) *Wrapper {
	if t.err != nil {
		return t
	}

	for _, f := range funcs {
		t.t = t.t.Funcs(f)
	}

	return t
}

func (t *Wrapper) FromFile(templateFile string) *Wrapper {
	if t.err != nil {
		return t
	}

	s, err := loadLocalOrEmbed(templateFile)
	if err != nil {
		t.err = errors.Wrapf(err, "error reading template: %s", templateFile)
	}

	return t.From(s)
}

func (t *Wrapper) From(s string) *Wrapper {
	if t.err != nil {
		return t
	}

	_, err := t.t.Parse(s)
	if err != nil {
		t.err = errors.Wrapf(err, "error parsing template")
	}

	return t
}

func (t *Wrapper) ToWriter(w io.Writer, data interface{}) error {
	if t.err != nil {
		return t.err
	}

	return t.t.Execute(w, data)
}

func (t *Wrapper) ToBytes(data interface{}) ([]byte, error) {
	if t.err != nil {
		return nil, t.err
	}

	buf := &bytes.Buffer{}

	err := t.ToWriter(buf, data)
	if err != nil {
		return nil, errors.Wrap(err, "error executing template")
	}

	return buf.Bytes(), nil
}

func (t *Wrapper) To(data interface{}) (string, error) {
	if t.err != nil {
		return "", t.err
	}

	b, err := t.ToBytes(data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (t *Wrapper) ToFile(file string, data interface{}) error {
	if t.err != nil {
		return t.err
	}

	if file == "stdout" {
		return t.ToWriter(os.Stdout, data)
	}

	if file == "stderr" {
		return t.ToWriter(os.Stderr, data)
	}

	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return errors.Wrap(err, "error creating output directory")
	}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	err = t.ToWriter(f, data)
	if closeErr := f.Close(); err == nil {
		return closeErr
	}

	return err
}
