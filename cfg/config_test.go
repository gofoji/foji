package cfg_test

import (
	"reflect"
	"testing"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/stringlist"
)

func TestLoad(t *testing.T) {
	type args struct {
		filename        string
		includeDefaults bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"invalid yaml", args{"test/invalid.yaml", false}, true},
		{"basic error", args{"test/DOESNOTEXIST.yaml", false}, true},
		{"basic", args{"test/basic.yaml", true}, false},
		{"basic merge", args{"test/sample.yaml", true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.Load(tt.args.filename, tt.args.includeDefaults)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_ = got
		})
	}
}

func TestProcesses_Keys(t *testing.T) {
	c, err := cfg.Load("test/sample.yaml", false)
	if err != nil {
		t.Errorf("Keys() Err = %v", err)
	}
	got := c.Processes.String()
	want := "badGroupTest,dbList,groupTest,openAPIDocs"
	if got != want {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestVersion(t *testing.T) {
	got := cfg.Version()
	if len(got) < 2 {
		t.Errorf("Version() = %v", got)
	}
}

func TestProcesses_Target(t *testing.T) {
	c, err := cfg.Load("test/sample.yaml", false)
	if err != nil {
		t.Errorf("Load() Err = %v", err)
	}

	tests := []struct {
		name    string
		targets []string
		want    int
		wantErr bool
	}{
		{"basic", []string{"dbList"}, 1, false},
		{"group", []string{"groupTest"}, 2, false},
		{"bad", []string{"bad"}, 0, true},
		{"badGroup", []string{"badGroupTest"}, 0, true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Processes.Target(tt.targets)
			if (err != nil) != tt.wantErr {
				t.Errorf("Target() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("Target() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParamMap_HasString(t *testing.T) {
	c, err := cfg.Load("test/sample.yaml", false)
	if err != nil {
		t.Errorf("Keys() Err = %v", err)
	}

	pp := c.Processes["openAPIDocs"].Params

	tests := []struct {
		name string
		want string
		ok   bool
	}{
		{"Package", "doc", true},
		{"NOPE", "", false},
		{"IntegerTest", "", false},
		{"FloatTest", "", false},
		{"Complex", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := pp.HasString(tt.name)
			if got != tt.want {
				t.Errorf("HasString() got = %v, want %v", got, tt.want)
			}

			if ok != tt.ok {
				t.Errorf("HasString() got1 = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestOutput_All(t *testing.T) {
	c, err := cfg.Load("test/sample.yaml", true)
	if err != nil {
		t.Errorf("Load() Err = %v", err)
	}

	tests := []struct {
		name string
		want stringlist.StringMap
	}{
		{"openAPIDocs", stringlist.StringMap{"!doc/handler.go": "foji/openapi/docs.go.tpl", "doc/embed_gen.go": "foji/embed.go.tpl"}},
		{"embed", stringlist.StringMap{}},
		{"groupTest", stringlist.StringMap{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.Processes[tt.name].All(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileInput_IsEmpty(t *testing.T) {
	c, err := cfg.Load("test/sample.yaml", true)
	if err != nil {
		t.Errorf("Load() Err = %v", err)
	}

	if c.Processes["openAPIDocs"].Files.IsEmpty() {
		t.Errorf("IsEmpty() returned true")
	}
}
