package sql

import (
	"fmt"
	"regexp"
)

var paramSearch = regexp.MustCompile(`@([[:alpha:]_][[:alnum:]_]+)(?:[\s)]|$)`)
var paramReplace = `(@%s)([\s)]|$)`
var paramFormat = `$$%d`

func sliceUniqMap(s [][]string, index int) []string {
	seen := make(map[string]struct{}, len(s))
	j := 0
	result := make([]string, len(s))

	for _, vv := range s {
		v := vv[index]
		if _, ok := seen[v]; ok {
			continue
		}

		seen[v] = struct{}{}
		result[j] = v
		j++
	}

	return result[:j]
}

func parseParams(sql string) (string, []string) {
	matches := sliceUniqMap(paramSearch.FindAllStringSubmatch(sql, -1), 1)
	for i, p := range matches {
		pr := regexp.MustCompile(fmt.Sprintf(paramReplace, p))
		sql = pr.ReplaceAllString(sql, fmt.Sprintf(paramFormat, i+1)+"$2")
	}

	return sql, matches
}
