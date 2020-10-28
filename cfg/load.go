package cfg

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gofoji/foji/embed"
	"gopkg.in/yaml.v3"
)

// Load reads a config file from filename.  If includeDefaults it will also load the defaults
// from the embedded config and merges.
func Load(filename string, includeDefaults bool) (Config, error) {
	c := Config{}

	f, err := os.Open(filename)
	if err != nil {
		return c, fmt.Errorf("can't open config file:%w", err)
	}

	b, err := ioutil.ReadAll(f)
	_ = f.Close() //nolint
	if err != nil {
		return c, fmt.Errorf("can't read config file:%w", err)
	}

	c, err = loadYaml(string(b))
	if err != nil {
		return c, fmt.Errorf("can't parse config file:%w", err)
	}

	if !includeDefaults {
		return c, nil
	}

	defaults, err := loadYaml(embed.FojiDotYaml)
	if err != nil {
		return c, fmt.Errorf("can't parse defaults:%w", err)
	}

	return c.Merge(defaults), nil
}

func loadYaml(source string) (Config, error) {
	c := Config{}

	d := yaml.NewDecoder(strings.NewReader(source))

	err := d.Decode(&c)
	if err != nil {
		return c, fmt.Errorf("can't decode yaml:%w", err)
	}

	return c, nil
}
