package cfg

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/gofoji/foji/embed"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func Load(filename string, includeDefaults bool) (Config, error) {
	c := Config{}

	f, err := os.Open(filename)
	if err != nil {
		return c, errors.Wrap(err, "can't open config file")
	}
	b, err := ioutil.ReadAll(f)
	_ = f.Close() //nolint
	if err != nil {
		return c, errors.Wrap(err, "can't read config file")
	}
	c, err = loadYaml(string(b))
	if err != nil {
		return c, errors.Wrap(err, "can't parse config file")
	}
	if !includeDefaults {
		return c, nil
	}

	defaults, err := loadYaml(embed.FojiDotYaml)
	if err != nil {
		return c, errors.Wrap(err, "can't parse defaults")
	}

	return c.Merge(defaults), nil
}

func loadYaml(source string) (Config, error) {
	c := Config{}

	d := yaml.NewDecoder(strings.NewReader(source))

	err := d.Decode(&c)
	if err != nil {
		return c, errors.Wrap(err, "can't read config file")
	}

	return c, nil
}
