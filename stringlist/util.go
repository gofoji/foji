package stringlist

import "regexp"

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

func contains(list []string, s string) bool {
	for x := range list {
		if s == list[x] {
			return true
		}
	}

	return false
}
