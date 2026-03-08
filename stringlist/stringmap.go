package stringlist

import (
	"maps"
	"slices"
)

type StringMap map[string]string

func (t StringMap) IsEmpty() bool {
	return len(t) == 0
}

func (t StringMap) Values() Strings {
	return slices.Collect(maps.Values(t))
}
