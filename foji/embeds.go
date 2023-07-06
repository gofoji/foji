package foji

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed *
var files embed.FS

const prefix = "foji/"

// Default reads and returns the content of the named file from the embeds.
func Default(name string) ([]byte, error) {
	return files.ReadFile(strings.TrimPrefix(name, prefix))
}

func walk(dir string) ([]string, error) {
	ff, err := files.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("readDir:%q:%w", dir, err)
	}

	currentPath := prefix
	if dir != "." {
		currentPath += dir + "/"
	}

	var out []string
	for _, f := range ff {
		if f.IsDir() {
			name := f.Name()
			if dir != "." {
				name = dir + "/" + name
			}

			dirFiles, err := walk(name)
			if err != nil {
				return nil, err
			}

			out = append(out, dirFiles...)

			continue
		}

		if !strings.HasSuffix(f.Name(), ".tpl") {
			continue
		}

		out = append(out, currentPath+f.Name())
	}

	return out, nil
}

func AllTemplates() ([]string, error) {
	return walk(".")
}

//go:embed foji.yaml
var DefaultConfig string

//go:embed init.yaml
var InitConfig []byte
