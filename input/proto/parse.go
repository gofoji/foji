package proto

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/gofoji/foji/input"
	"github.com/sirupsen/logrus"
	"github.com/yoheimuta/go-protoparser"
	"github.com/yoheimuta/go-protoparser/parser"
)

type (
	Messages []*Message

	Message struct {
		*parser.Message
		Fields []*Field
	}

	Field struct {
		*parser.Field
		Message *Message
	}

	Enums []*Enum

	Enum struct {
		*parser.Enum
		Fields []*EnumField
	}

	EnumField struct {
		*parser.EnumField
	}

	PBFile struct {
		Messages    Messages
		Enums       Enums
		lastMessage *Message
		lastEnum    *Enum
	}

	PBFileGroup  []PBFile
	PBFileGroups []PBFileGroup
)

func (f *Field) OptionByName(name string) string {
	for _, o := range f.FieldOptions {
		if o.OptionName == name {
			s, err := strconv.Unquote(o.Constant)
			if err != nil {
				return o.Constant
			}

			return s
		}
	}

	return ""
}

func (ee Enums) ByName(name string) *Enum {
	for _, e := range ee {
		if e.EnumName == name {
			return e
		}
	}

	return nil
}

func (mm Messages) ByName(name string) *Message {
	for _, m := range mm {
		if m.MessageName == name {
			return m
		}
	}

	return nil
}

func (f *Field) Path() string {
	if f.Message != nil {
		return f.Message.MessageName + "." + f.FieldName
	}

	return f.FieldName
}

func Parse(_ context.Context, logger logrus.FieldLogger, inGroups []input.FileGroup) (PBFileGroups, error) {
	result := make(PBFileGroups, len(inGroups))

	for i, ff := range inGroups {
		group := make(PBFileGroup, len(ff.Files))

		for i, f := range ff.Files {
			logger.WithField("filename", f.Source).Debug("Parsing Proto")

			r, err := protoparser.Parse(bytes.NewReader(f.Content))
			if err != nil {
				return nil, fmt.Errorf("proto.Parse:%w", err)
			}

			d := PBFile{}
			r.Accept(&d)

			group[i] = d
		}

		result[i] = group
	}

	return result, nil
}

func (d *PBFile) VisitEnum(n *parser.Enum) (next bool) {
	e := &Enum{Enum: n}
	d.Enums = append(d.Enums, e)
	d.lastEnum = e

	return true
}

func (d *PBFile) VisitEnumField(n *parser.EnumField) (next bool) {
	e := &EnumField{n}
	d.lastEnum.Fields = append(d.lastEnum.Fields, e)

	return true
}

func (d *PBFile) VisitField(n *parser.Field) (next bool) {
	f := &Field{Field: n}
	d.lastMessage.Fields = append(d.lastMessage.Fields, f)

	return true
}

func (d *PBFile) VisitMessage(n *parser.Message) (next bool) {
	m := &Message{Message: n}
	d.Messages = append(d.Messages, m)
	d.lastMessage = m

	return true
}

// The rest of these are required by the visitor interface.
func (d *PBFile) VisitExtend(n *parser.Extend) (next bool) {
	return true
}

func (d *PBFile) VisitExtensions(n *parser.Extensions) (next bool) {
	return true
}

func (d *PBFile) VisitGroupField(n *parser.GroupField) (next bool) {
	return true
}

func (d *PBFile) VisitImport(n *parser.Import) (next bool) {
	return true
}

func (d *PBFile) VisitMapField(n *parser.MapField) (next bool) {
	return true
}

func (d *PBFile) VisitComment(n *parser.Comment) {}

func (d *PBFile) VisitEmptyStatement(n *parser.EmptyStatement) (next bool) {
	return true
}

func (d *PBFile) VisitOneof(n *parser.Oneof) (next bool) {
	return true
}

func (d *PBFile) VisitOneofField(n *parser.OneofField) (next bool) {
	return true
}

func (d *PBFile) VisitOption(n *parser.Option) (next bool) {
	return true
}

func (d *PBFile) VisitPackage(n *parser.Package) (next bool) {
	return true
}

func (d *PBFile) VisitReserved(n *parser.Reserved) (next bool) {
	return true
}

func (d *PBFile) VisitRPC(n *parser.RPC) (next bool) {
	return true
}

func (d *PBFile) VisitService(n *parser.Service) (next bool) {
	return true
}

func (d *PBFile) VisitSyntax(n *parser.Syntax) (next bool) {
	return true
}
