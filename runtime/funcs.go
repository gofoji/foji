package runtime

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/codemodus/kace"
	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/color"
	"github.com/gofoji/foji/stringlist"
	"github.com/jinzhu/inflection"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

var ErrRuntime = Error("runtime")

var Funcs = map[string]interface{}{
	// Case
	"camel":      kace.Camel,
	"kebab":      kace.Kebab,
	"kebabUpper": kace.KebabUpper,
	"pascal":     kace.Pascal,
	"snake":      kace.Snake,
	"snakeUpper": kace.SnakeUpper,
	"unchanged":  Unchanged,

	// Strings
	"fields":       strings.Fields,
	"indexAny":     strings.IndexAny,
	"lastIndex":    strings.LastIndex,
	"lastIndexAny": strings.LastIndexAny,
	"replaceAll":   strings.ReplaceAll,
	"strIndex":     strings.Index,
	"trimLeft":     strings.TrimLeft,
	"trimRight":    strings.TrimRight,

	// Inflection
	"pluralName":       inflection.Plural,
	"pluralUniqueName": PluralUnique,
	"singular":         inflection.Singular,

	// General
	"pad":         Pad,
	"notEmpty":    NotEmpty,
	"version":     cfg.Version,
	"toSlice":     ToSlice,
	"numbers":     Numbers,
	"sum":         Sum,
	"inc":         Inc,
	"replaceEach": ReplaceEach,
	"fileWithExt": FileWithExt,
	"isNil":       IsNil,
	"isNotNil":    IsNotNil,
	"in":          In,

	// Go Special
	"backQuote": BackQuote,
	"csv":       Csv,
	"goToken":   GoToken,
	"goDoc":     GoDoc,

	// Console
	"blue":       color.Blue,
	"color":      color.ByName,
	"colors":     color.ByNames,
	"colorReset": color.Clear,
	"cyan":       color.Cyan,
	"green":      color.Green,
	"magenta":    color.Magenta,
	"red":        color.Red,
	"yellow":     color.Yellow,
}

// Token converts the in string into a valid Go Token by converting "/" an ".".
func GoToken(in string) string {
	s := strings.ReplaceAll(in, "/", "_SLASH_")
	s = strings.ReplaceAll(s, ".", "_DOT_")
	s = strings.ReplaceAll(s, "$", "_DOLLAR_")

	return s
}

func BackQuote(in string) string {
	return strings.ReplaceAll(in, "`", "`+\"`\"+`")
}

func CaseFuncs(name string) map[string]interface{} {
	if name == "" {
		name = "unchanged"
	}

	return map[string]interface{}{"case": Case(name), "cases": Cases(name)}
}

func Case(name string) interface{} {
	return Funcs[name]
}

func Cases(name string) interface{} {
	m, ok := Funcs[name].(stringlist.StringMapper)
	if ok {
		return stringlist.Mapper(m)
	}

	return nil
}

const PluralSuffix = "List"

// PluralUnique guarantees a unique name for a Plural of the input.
func PluralUnique(s string) string {
	result := inflection.Plural(s)
	if result == s {
		return s + PluralSuffix
	}

	return result
}

// Unchanged is used as a Case function that does not alter the string.
func Unchanged(s string) string {
	return s
}

func Pad(s string, size int) string {
	return fmt.Sprintf("%-"+strconv.Itoa(size)+"s", s)
}

// ToSlice returns the arguments as a single slice.  If all the arguments are
// strings, they are returned as a stringlist.Strings, otherwise they're returned as
// []interface{}.
func ToSlice(vv ...interface{}) interface{} {
	ss := make(stringlist.Strings, len(vv))

	for x := range vv {
		if s, ok := vv[x].(string); ok {
			ss[x] = s
		} else {
			// something was not a string, so just return the []interface{}
			return vv
		}
	}

	return ss
}

// numbers returns a slice of strings of the numbers start to end (inclusive).
func Numbers(start, end int) stringlist.Strings {
	var ss stringlist.Strings

	for x := start; x <= end; x++ {
		ss = append(ss, strconv.Itoa(x))
	}

	return ss
}

// Sum returns the sum of its arguments.
func Sum(vv ...int) int {
	x := 0

	for _, v := range vv {
		x += v
	}

	return x
}

// Inc increments the argument's value by 1.
func Inc(x int) int {
	return x + 1
}

func ReplaceEach(s, new string, olds ...string) string {
	for _, old := range olds {
		s = strings.ReplaceAll(s, old, new)
	}

	return s
}

func FileWithExt(path, ext string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + ext
}

func Csv(in stringlist.Strings) string {
	return in.Join(", ")
}

func NotEmpty(in string) bool {
	return len(in) > 0
}

// IsNil returns true if the input or referenced object is nil.
func IsNil(i interface{}) bool {
	return (*[2]uintptr)(unsafe.Pointer(&i))[1] == 0
}

// IsNotNil returns false if the input or referenced object is nil.
func IsNotNil(i interface{}) bool {
	return !IsNil(i)
}

func In(needle interface{}, haystack ...interface{}) (bool, error) {
	if haystack == nil {
		return false, nil
	}

	tp := reflect.TypeOf(haystack).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		var item interface{}

		l2 := reflect.ValueOf(haystack)

		l := l2.Len()
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if reflect.DeepEqual(needle, item) {
				return true, nil
			}
		}

		return false, nil
	default:
		return false, fmt.Errorf("%w: must be iterable type, found type %s", ErrRuntime, tp)
	}
}

const (
	CommentPrefix = "//"
	MaxWidth      = 80
)

// GoDoc wraps the string to a MaxWidth and prepends with CommentPrefix.
func GoDoc(s string) string {
	ss := strings.Split(s, "\n")
	out := CommentPrefix
	length := 0

	for lineNumber, s := range ss {
		ll := strings.Split(s, " ")
		for _, l := range ll {
			if len(l)+length > MaxWidth {
				out += "\n" + CommentPrefix
				length = 0
			}

			length += len(l)
			out += " " + l
		}

		if lineNumber > 0 {
			out += "\n" + CommentPrefix
		}
	}

	return out
}
