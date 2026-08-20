package cfg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/gofoji/foji/cfg"
)

// declared in an order that is neither alphabetical nor reverse alphabetical,
// so a map-ordering regression cannot pass by luck.
const orderedSource = `
zebra: 1
apple: 2
mango: 3
banana: 4
cherry: 5
apricot: 6
durian: 7
elderberry: 8
`

var orderedKeys = []string{"zebra", "apple", "mango", "banana", "cherry", "apricot", "durian", "elderberry"}

func TestOrderedMap_UnmarshalPreservesDeclarationOrder(t *testing.T) {
	var m cfg.OrderedMap

	require.NoError(t, yaml.Unmarshal([]byte(orderedSource), &m))

	assert.Equal(t, orderedKeys, []string(m.Keys()))
	assert.Equal(t, len(orderedKeys), m.Len())
	assert.Equal(t, "3", m.Value("mango"))
	assert.True(t, m.Has("durian"))
	assert.False(t, m.Has("nope"))
	assert.Empty(t, m.Value("nope"))
}

func TestOrderedMap_EntriesIteratesInOrder(t *testing.T) {
	var m cfg.OrderedMap

	require.NoError(t, yaml.Unmarshal([]byte(orderedSource), &m))

	var got []string
	for key := range m.Entries() {
		got = append(got, key)
	}

	assert.Equal(t, orderedKeys, got)
}

func TestOrderedMap_RoundTripPreservesOrder(t *testing.T) {
	var m cfg.OrderedMap

	require.NoError(t, yaml.Unmarshal([]byte(orderedSource), &m))

	out, err := yaml.Marshal(m)
	require.NoError(t, err)

	var back cfg.OrderedMap

	require.NoError(t, yaml.Unmarshal(out, &back))
	assert.Equal(t, orderedKeys, []string(back.Keys()))
	assert.Equal(t, m.Values(), back.Values())
}

func TestOrderedMap_SetKeepsPositionOnOverwrite(t *testing.T) {
	m := cfg.NewOrderedMap("a", "1", "b", "2", "c", "3")

	m.Set("a", "replaced")

	assert.Equal(t, []string{"a", "b", "c"}, []string(m.Keys()))
	assert.Equal(t, "replaced", m.Value("a"))
	assert.Equal(t, 3, m.Len())
}

func TestOrderedMap_SetAllAppendsNewKeysInOrder(t *testing.T) {
	m := cfg.NewOrderedMap("a", "1", "b", "2")

	m.SetAll(cfg.NewOrderedMap("b", "replaced", "c", "3"))

	assert.Equal(t, []string{"a", "b", "c"}, []string(m.Keys()))
	assert.Equal(t, []string{"1", "replaced", "3"}, []string(m.Values()))
}

func TestOrderedMap_ZeroValueIsUsable(t *testing.T) {
	var m cfg.OrderedMap

	assert.Zero(t, m.Len())
	assert.True(t, m.IsZero())
	assert.False(t, m.Has("a"))
	assert.Empty(t, m.Value("a"))
	assert.Empty(t, m.Keys())

	for range m.Entries() {
		t.Fatal("zero value yielded an entry")
	}

	m.Set("a", "1")
	assert.Equal(t, 1, m.Len())
	assert.False(t, m.IsZero())
}

func TestOrderedMap_RejectsNonMapping(t *testing.T) {
	var m cfg.OrderedMap

	assert.Error(t, yaml.Unmarshal([]byte("- a\n- b\n"), &m))
}

func TestNewOrderedMap_PanicsOnOddArgs(t *testing.T) {
	assert.Panics(t, func() { cfg.NewOrderedMap("a") })
}

// The output templates are declared inline on the process, so this covers the
// path that actually decides weld generation order.
func TestOutput_PreservesTemplateOrder(t *testing.T) {
	const source = `
processes:
  demo:
    OpenAPIFile:
      zzz_last.go: last.tpl
      aaa_first.go: first.tpl
      mmm_middle.go: middle.tpl
`

	c, err := cfg.LoadYaml(source)
	require.NoError(t, err)

	got := c.Processes["demo"].Output["OpenAPIFile"]
	assert.Equal(t, []string{"zzz_last.go", "aaa_first.go", "mmm_middle.go"}, []string(got.Keys()))
}

func TestOutput_AllIsStableAcrossKinds(t *testing.T) {
	const source = `
processes:
  demo:
    OpenAPIFile:
      b.go: b.tpl
      a.go: a.tpl
    DbTable:
      d.go: d.tpl
      c.go: c.tpl
`

	c, err := cfg.LoadYaml(source)
	require.NoError(t, err)

	// Output kinds are visited sorted (DbTable before OpenAPIFile); within a
	// kind the declared order is kept.
	want := []string{"d.go", "c.go", "b.go", "a.go"}

	for range 10 {
		assert.Equal(t, want, []string(c.Processes["demo"].All().Keys()))
	}
}
