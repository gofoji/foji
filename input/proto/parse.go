package proto

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofoji/foji/input"
	"github.com/sirupsen/logrus"
	"github.com/yoheimuta/go-protoparser"
	"github.com/yoheimuta/go-protoparser/parser"
)

type Messages []*Message

type Message struct {
	*parser.Message
	Fields []*Field
}
type Field struct {
	*parser.Field
	Message *Message
}

func (f *Field) OptionByName(name string) string {
	for _, o := range f.FieldOptions{
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

type Enum struct {
	*parser.Enum
	Fields []*EnumField
}
type EnumField struct {
	*parser.EnumField
}

type Enums []*Enum

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

type PBFile struct {
	Messages    Messages
	Enums       Enums
	lastMessage *Message
	lastEnum    *Enum
}

type PBFileGroup []PBFile
type PBFileGroups []PBFileGroup

func (f *Field) Path() string {
	if f.Message != nil {
		return f.Message.MessageName + "." + f.FieldName
	}
	return f.FieldName
}

func Parse(ctx context.Context, logger logrus.FieldLogger, inGroups []input.FileGroup) (PBFileGroups, error) {
	var result PBFileGroups

	for _, ff := range inGroups {
		var group PBFileGroup
		for _, f := range ff.Files {
			r, err := protoparser.Parse(strings.NewReader(f.Content))
			if err != nil {
				return nil, err
			}

			d := PBFile{}
			r.Accept(&d)

			group = append(group, d)
		}
		result = append(result, group)
	}

	return result, nil
}

func (d *PBFile) VisitComment(n *parser.Comment) {
}
func (d *PBFile) VisitEmptyStatement(n *parser.EmptyStatement) (next bool) {
	return true
}
func (d *PBFile) VisitEnum(n *parser.Enum) (next bool) {
	e := &Enum{
		Enum: n,
	}
	d.Enums = append(d.Enums, e)
	d.lastEnum = e
	return true
}
func (d *PBFile) VisitEnumField(n *parser.EnumField) (next bool) {
	e := &EnumField{n}
	d.lastEnum.Fields = append(d.lastEnum.Fields, e)
	return true
}
func (d *PBFile) VisitExtend(n *parser.Extend) (next bool) {
	return true
}
func (d *PBFile) VisitExtensions(n *parser.Extensions) (next bool) {
	return true
}
func (d *PBFile) VisitField(n *parser.Field) (next bool) {
	f := &Field{
		Field: n,
	}
	d.lastMessage.Fields = append(d.lastMessage.Fields, f)
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
func (d *PBFile) VisitMessage(n *parser.Message) (next bool) {
	m := &Message{
		Message: n,
	}
	d.Messages = append(d.Messages, m)
	d.lastMessage = m
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
