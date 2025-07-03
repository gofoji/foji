package output

import (
	"strings"

	"github.com/rs/zerolog"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input/proto"
)

const (
	ProtoAll       = "ProtoAll"
	ProtoFileGroup = "ProtoFileGroup"
	ProtoFile      = "ProtoFile"
)

func HasProtoOutput(o cfg.Output) bool {
	return hasAnyOutput(o, ProtoAll, ProtoFileGroup, ProtoFile)
}

func Proto(p cfg.Process, fn cfg.FileHandler, l zerolog.Logger, groups proto.PBFileGroups, simulate bool) error {
	base := ProtoContext{
		Context:    Context{Process: p, Logger: l},
		FileGroups: groups,
	}
	runner := NewProcessRunner(p.RootDir, fn, l, simulate)

	err := runner.process(p.Output[ProtoAll], &base)
	if err != nil {
		return err
	}

	for _, ff := range groups {
		ctx := ProtoFileGroupContext{
			ProtoContext: base,
			FileGroup:    ff,
		}

		err := runner.process(p.Output[ProtoFileGroup], &ctx)
		if err != nil {
			return err
		}

		for _, f := range ff {
			ctx := ProtoFileContext{
				ProtoFileGroupContext: ctx,
				PBFile:                f,
			}

			err := runner.process(p.Output[ProtoFile], &ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type ProtoContext struct {
	Context
	Imports

	FileGroups proto.PBFileGroups
}

type ProtoFileGroupContext struct {
	ProtoContext

	FileGroup proto.PBFileGroup
}

type ProtoFileContext struct {
	ProtoFileGroupContext

	proto.PBFile
}

func (q ProtoContext) IsEnum(name string) bool {
	for _, g := range q.FileGroups {
		for _, f := range g {
			e := f.Enums.ByName(name)
			if e != nil {
				return true
			}
		}
	}

	return false
}

func (q ProtoContext) IsMessage(name string) bool {
	for _, g := range q.FileGroups {
		for _, f := range g {
			e := f.Messages.ByName(name)
			if e != nil {
				return true
			}
		}
	}

	return false
}

func (q ProtoContext) HasMessage(msg *proto.Message) bool {
	for _, f := range msg.Fields {
		if q.IsMessage(f.Type) {
			return true
		}
	}

	return false
}

func (q ProtoContext) GetType(f proto.Field, pkg string) string {
	pp := strings.Split(f.Path(), ".")
	for i := range pp {
		p := strings.Join(pp[i:], ".")

		t, ok := q.Maps.Type["."+p]
		if ok {
			return q.CheckPackage(t, pkg)
		}
	}

	t, ok := q.Maps.Type[f.Type]
	if ok {
		return q.CheckPackage(t, pkg)
	}

	// TODO Valid assumption for type reference?
	// If not in the above mappings, then assume it is a reference to another Message in the package
	if q.IsEnum(f.Type) {
		return f.Type
	}

	return "*" + f.Type
}
