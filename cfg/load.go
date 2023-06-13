package cfg

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gofoji/foji/embed"
)

// Load reads a config file from filename.  If includeDefaults it will also load the defaults
// from the embedded config and merges.
func Load(filename string, includeDefaults bool) (Config, error) {
	cfg := Config{}

	f, err := os.Open(filename)
	if err != nil {
		return cfg, fmt.Errorf("can't open config file:%w", err)
	}

	b, err := io.ReadAll(f)
	_ = f.Close() //nolint
	if err != nil {
		return cfg, fmt.Errorf("can't read config file:%w", err)
	}

	cfg, err = LoadYaml(string(b))
	if err != nil {
		return cfg, fmt.Errorf("can't parse config file:%w", err)
	}

	if !includeDefaults {
		return cfg, nil
	}

	defaults, err := LoadYaml(embed.FojiDotYaml)
	if err != nil {
		return cfg, fmt.Errorf("can't parse defaults:%w", err)
	}

	return cfg.Merge(defaults), nil
}

func LoadYaml(source string) (Config, error) {
	c := Config{}

	d := yaml.NewDecoder(strings.NewReader(source))

	if err := d.Decode(&c); err != nil {
		return c, fmt.Errorf("can't decode yaml:%w", err)
	}

	return c, nil
}
