package cfg

import (
	"fmt"
	"iter"
	"slices"

	"go.yaml.in/yaml/v4"

	"github.com/gofoji/foji/errs"
	"github.com/gofoji/foji/stringlist"
)

const errOrderedMap = errs.Error("OrderedMap")

// OrderedMap is a string to string map that remembers the order its keys were
// declared in the config file.
//
// Config maps whose iteration order is observable in a weld use this instead of
// a plain Go map, because Go randomizes map iteration:
//
//   - the template list under each output kind decides the order files are
//     generated, and templates share accumulated context state (see Imports),
//   - file rewrite rules are first-match-wins.
//
// Both would otherwise vary between runs of an unchanged config. Declaration
// order is also the order a reader expects from the file they wrote.
//
// The zero value is an empty map and is ready to use.
//
// Mutators take a pointer; readers take a value so that they still work on a
// map element such as `p.Output[DbTable].Len()`, which is not addressable.
//
//nolint:recvcheck
type OrderedMap struct {
	keys   []string
	values map[string]string
}

// NewOrderedMap builds a map from alternating key, value arguments. It panics
// on an odd number of arguments, so it is only for literals known at compile
// time; config is loaded through UnmarshalYAML instead.
func NewOrderedMap(pairs ...string) OrderedMap {
	if len(pairs)%2 != 0 {
		panic("cfg.NewOrderedMap: odd number of arguments")
	}

	var out OrderedMap

	for i := 0; i < len(pairs); i += 2 {
		out.Set(pairs[i], pairs[i+1])
	}

	return out
}

// Set adds or replaces a value. Replacing keeps the key in its original
// position, so a merge never reorders entries that were already declared.
func (m *OrderedMap) Set(key, value string) {
	if m.values == nil {
		m.values = map[string]string{}
	}

	if _, ok := m.values[key]; !ok {
		m.keys = append(m.keys, key)
	}

	m.values[key] = value
}

// SetAll copies every entry of from, in from's order.
func (m *OrderedMap) SetAll(from OrderedMap) {
	for key, value := range from.Entries() {
		m.Set(key, value)
	}
}

// Len returns the number of entries.
func (m OrderedMap) Len() int {
	return len(m.keys)
}

// IsZero reports whether the map is empty. It implements yaml.IsZeroer so that
// `omitempty` drops an empty map on encode.
func (m OrderedMap) IsZero() bool {
	return len(m.keys) == 0
}

// Has reports whether key is present.
func (m OrderedMap) Has(key string) bool {
	_, ok := m.values[key]

	return ok
}

// Value returns the value for key, or the empty string if it is absent.
func (m OrderedMap) Value(key string) string {
	return m.values[key]
}

// Keys returns the keys in declaration order.
func (m OrderedMap) Keys() stringlist.Strings {
	return stringlist.Strings(slices.Clone(m.keys))
}

// Values returns the values in declaration order.
func (m OrderedMap) Values() stringlist.Strings {
	out := make(stringlist.Strings, len(m.keys))
	for i, key := range m.keys {
		out[i] = m.values[key]
	}

	return out
}

// Entries iterates the map in declaration order.
func (m OrderedMap) Entries() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, key := range m.keys {
			if !yield(key, m.values[key]) {
				return
			}
		}
	}
}

// UnmarshalYAML reads a mapping node directly rather than decoding into a Go
// map, because a mapping node's contents are in document order.
func (m *OrderedMap) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("%w: expected a mapping at line %d", errOrderedMap, node.Line)
	}

	*m = OrderedMap{}

	// A mapping node holds its keys and values alternating in Content.
	for i := 0; i+1 < len(node.Content); i += 2 {
		var key string

		err := node.Content[i].Decode(&key)
		if err != nil {
			return fmt.Errorf("%w: decoding key at line %d: %w", errOrderedMap, node.Content[i].Line, err)
		}

		var value string

		err = node.Content[i+1].Decode(&value)
		if err != nil {
			return fmt.Errorf("%w: decoding value for %q: %w", errOrderedMap, key, err)
		}

		m.Set(key, value)
	}

	return nil
}

// MarshalYAML emits a mapping node so that dumpConfig round-trips the declared
// order instead of re-sorting it.
func (m OrderedMap) MarshalYAML() (any, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}

	for _, key := range m.keys {
		keyNode := &yaml.Node{}

		err := keyNode.Encode(key)
		if err != nil {
			return nil, fmt.Errorf("%w: encoding key %q: %w", errOrderedMap, key, err)
		}

		valueNode := &yaml.Node{}

		err = valueNode.Encode(m.values[key])
		if err != nil {
			return nil, fmt.Errorf("%w: encoding value for %q: %w", errOrderedMap, key, err)
		}

		node.Content = append(node.Content, keyNode, valueNode)
	}

	return node, nil
}
