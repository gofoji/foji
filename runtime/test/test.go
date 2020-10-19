package main

import (
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/codemodus/kace"
	"github.com/gofoji/foji/runtime"
	"github.com/gofoji/foji/tpl"
)

type c func(lists ...interface{}) interface{}

func main() {
	x := runtime.Funcs
	z := sprig.GenericFuncMap()
	for key, _ := range x {
		if _, ok := z[key]; ok {
			println(key)
		}
	}
	println(kace.Pascal("asdf asdf asdf"))
	println(strings.Title("asdf asdf asdf"))

	t := tpl.New("test")
	r, err := t.Funcs(sprig.GenericFuncMap()).From(`{{list "asdf" "a" "b" "c" | join "" }}`).To(nil)
	if err != nil {
		panic(err)
	}
	println(r)
}
