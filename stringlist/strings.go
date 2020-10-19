package stringlist

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/codemodus/kace"
)

type StringMap map[string]string
type Strings []string
type Strings2D []Strings
type StringMapper = func(string) string
type StringsMapper = func(Strings) Strings
type FilterFunc = func(string) bool

// Sprintf calls fmt.Sprintf(format, str) for every string in this value and
// returns the results as a new Strings.
func (s Strings) Sprintf(format string) Strings {
	ret := make(Strings, len(s))
	for x := range s {
		ret[x] = fmt.Sprintf(format, s[x])
	}

	return ret
}

func Mapper(f StringMapper) StringsMapper {
	return func(ss Strings) Strings {
		ret := make(Strings, len(ss))
		for x := range ss {
			ret[x] = f(ss[x])
		}

		return ret
	}
}

func MatchFilters(filters Strings, s string) bool {
	for _, f := range filters {
		match, _ := regexp.MatchString(f, s)
		if match {
			return true
		}
	}

	return false
}

func (s Strings) Filter(fn FilterFunc) Strings {
	ret := make(Strings, 0, len(s))

	for _, t := range s {
		if fn(t) {
			ret = append(ret, t)
		}
	}

	return ret
}

func (s Strings) Filters(filters Strings) Strings {
	ret := make(Strings, 0, len(s))

	for _, t := range s {
		if !MatchFilters(filters, t) {
			ret = append(ret, t)
		}
	}

	return ret
}

// AnyMatches treats each string as a regexp and returns true if any match the input string
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
	ret := make(Strings, 0, len(s))
	for x := range s {
		ret = append(ret, kace.Camel(s[x]))
	}

	return ret
}

// Pascal applies kace.Pascal to each string.
func (s Strings) Pascal() Strings {
	ret := make(Strings, 0, len(s))
	for x := range s {
		ret = append(ret, kace.Pascal(s[x]))
	}

	return ret
}

// Contains returns true if the param exists in the Strings.
func (s Strings) Contains(v string) bool {
	return contains(s, v)
}

// ContainsAny returns true if any of the the params exists in the Strings.
func (s Strings) ContainsAny(v Strings) bool {
	for _, x := range v {
		if s.Contains(x) {
			return true
		}
	}

	return false
}

// ContainsAll returns true if all of the the params exists in the Strings.
func (s Strings) ContainsAll(v Strings) bool {
	for _, x := range v {
		if !s.Contains(x) {
			return false
		}
	}

	return true
}
func contains(list []string, s string) bool {
	for x := range list {
		if s == list[x] {
			return true
		}
	}

	return false
}

// Max returns the length of the longest string
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

// Sort returns a copy of the list sorted
func (s Strings) Sort() Strings {
	result := make(Strings, len(s))
	for x := range s {
		result[x] = s[x]
	}
	sort.Strings(result)
	return result
}

func (t StringMap) IsEmpty() bool {
	return len(t) == 0
}

func (t StringMap) Values() Strings {
	result := Strings{}
	for _, v := range t {
		result = append(result, v)
	}
	return result
}
