package stringlist

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/codemodus/kace"
)

type (
	Strings       []string
	Strings2D     []Strings
	StringMapper  = func(string) string
	StringsMapper = func(Strings) Strings
	FilterFunc    = func(string) bool
)

// Sprintf calls fmt.Sprintf(format, str) for every string in this value and
// returns the results as a new Strings.
func (s Strings) Sprintf(format string) Strings {
	ret := make(Strings, len(s))
	for x := range s {
		ret[x] = fmt.Sprintf(format, s[x])
	}

	return ret
}

// Filter applies the given filter function fn and returns a copy of the list for each string that returned true.
func (s Strings) Filter(fn FilterFunc) Strings {
	ret := make(Strings, 0, len(s))

	for _, t := range s {
		if fn(t) {
			ret = append(ret, t)
		}
	}

	return ret
}

// Filters returns the list of strings that does not match any of the regex filters.
func (s Strings) Filters(filters Strings) Strings {
	ret := make(Strings, 0, len(s))

	for _, t := range s {
		if !MatchFilters(filters, t) {
			ret = append(ret, t)
		}
	}

	return ret
}

// AnyMatches treats each string as a regexp and returns true if any match the input string.
func (s Strings) AnyMatches(in string) bool {
	for _, f := range s {
		match, _ := regexp.MatchString(f, in)
		if match {
			return true
		}
	}

	return false
}

// Join concatenates the elements of its first argument to create a single string. The separator
// string sep is placed between elements in the resulting string.
func (s Strings) Join(sep string) string {
	return strings.Join(s, sep)
}

// Camel applies kace.Camel to each string.
func (s Strings) Camel() Strings {
	return s.Map(kace.Camel)
}

// Pascal applies kace.Pascal to each string.
func (s Strings) Pascal() Strings {
	return s.Map(kace.Pascal)
}

// Map apples the transform function f to each element and returns the mapped list.
func (s Strings) Map(f StringMapper) Strings {
	ret := make(Strings, 0, len(s))
	for x := range s {
		ret = append(ret, f(s[x]))
	}

	return ret
}

// Contains returns true if the param exists in the Strings.
func (s Strings) Contains(v string) bool {
	return contains(s, v)
}

// ContainsAny returns true if any of the params exists in the Strings.
func (s Strings) ContainsAny(v Strings) bool {
	for _, x := range v {
		if s.Contains(x) {
			return true
		}
	}

	return false
}

// ContainsAll returns true if all the params exists in the Strings.
func (s Strings) ContainsAll(v Strings) bool {
	for _, x := range v {
		if !s.Contains(x) {
			return false
		}
	}

	return true
}

// Max returns the length of the longest string.
func (s Strings) Max() int {
	result := 0

	for _, x := range s {
		l := len(x)
		if l > result {
			result = l
		}
	}

	return result
}

// Sort returns a copy of the list sorted.
func (s Strings) Sort() Strings {
	result := make(Strings, len(s))
	copy(result, s)
	sort.Strings(result)

	return result
}
