package sql

import (
	"context"
	"regexp"
	"strings"

	"github.com/gofoji/foji/input"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var commentRE = regexp.MustCompile(`^--[-]*`)
var paramRE = regexp.MustCompile(`^@`)
var nameRE = regexp.MustCompile(`^.`)

type Repo interface {
	DescribeQuery(ctx context.Context, q *Query) error
}

type Parser struct {
	repo   Repo
	logger logrus.FieldLogger
	ctx    context.Context
}

type File struct {
	input.File
	Queries []Query
}

type FileGroup []File

type FileGroups []FileGroup

type Query struct {
	Filename string
	Name     string
	Result   Result
	Params   Params
	SQL      string
	Raw      string
	Comment  string
	Type     string
}

type Result struct {
	Type          string // Type of the query result, provided in query descriptor or inferred from table typing
	IsSingleTable bool   // If a single table in the response, you can use table type mapping to resolve type
	Schema        string `json:"schema,omitempty"` // Schema of the single table response, empty otherwise
	Table         string `json:"table,omitempty"`  // Table of the single table response, empty otherwise
	Params        Params `json:"params,omitempty"` // Params in the response, empty otherwise
}

func (r Result) TypeParam() *Param {
	return &Param{
		Type:      r.Type,
		Generated: r.GenerateType(),
	}
}

func (r Result) GenerateType() bool {
	return !strings.Contains(r.Type, ".")
}

func (q Query) IsType(t string) bool {
	return q.Type == t
}

func (p Parser) readDescriptor(s string) Query {
	q := Query{Raw: s, Type: "query"}
	ss := strings.Split(s, "\n")

	for _, line := range ss {
		if commentRE.MatchString(line) {
			p.parseDescriptorLine(line, &q)
		}
	}

	return q
}

func (p Parser) parseDescriptorLine(line string, q *Query) {
	l := strings.TrimSpace(commentRE.ReplaceAllString(line, ""))
	if paramRE.MatchString(l) {
		p.parseDescriptorParam(l, q)
	} else if nameRE.MatchString(l) {
		p.parseDescriptorHeader(l, q)
	}
}

func (p Parser) parseDescriptorHeader(l string, q *Query) {
	/*
		Example headers:
		-- .FuncName
		-- .FuncName ResultType
		-- .FuncName ResultType QueryType
	*/
	ll := strings.Split(strings.TrimSpace(nameRE.ReplaceAllString(l, "")), " ")
	if len(ll) > 0 {
		q.Name = ll[0]
	}

	if len(ll) > 1 {
		q.Result.Type = ll[1]
	}

	if len(ll) > 2 {
		q.Type = ll[2]
	}

	p.logger.WithField("name", q.Name).Debug("Query Info")
}

func (p Parser) parseDescriptorParam(l string, f *Query) {
	pp := strings.Split(l, " ")
	if len(pp) == 1 {
		param := Param{
			Ordinal: len(f.Params),
			Name:    l,
			Query:   f,
		}
		f.Params = append(f.Params, &param)

		p.logger.WithField("param", p).Debug("Query Param")
	} else {
		param := Param{
			Ordinal: len(f.Params),
			Name:    pp[0],
			Type:    pp[1],
			Query:   f,
		}
		f.Params = append(f.Params, &param)

		p.logger.WithField("param", p).Debug("Query Param with Type")
	}
}

func Parse(ctx context.Context, logger logrus.FieldLogger, repo Repo, inGroups []input.FileGroup) (FileGroups, error) {
	var result FileGroups

	p := Parser{
		repo:   repo,
		logger: logger,
		ctx:    ctx,
	}

	for _, source := range inGroups {
		var fileSet FileGroup
		for _, f := range source.Files {
			logger.WithField("filename", f.Name).Debug("Parsing")

			resultFile := File{File: f}
			statements := strings.Split(f.Content, ";")
			for _, s := range statements {
				s := strings.TrimSpace(s)
				if s == "" {
					continue
				}

				logger.WithField("sql", s).Trace("Parsing")

				q := p.readDescriptor(s)
				q.Filename = f.Name

				s, params := parseParams(s)
				for a, b := range params {
					name := b
					p := q.Params.ByName(name)

					if p == nil {
						p = &Param{
							Ordinal: len(q.Params),
							Name:    name,
							Type:    "string",
							Query:   &q,
						}
						q.Params = append(q.Params, p)
					}

					p.QueryPosition = a
				}

				q.SQL = s

				err := p.repo.DescribeQuery(context.Background(), &q)
				if err != nil {
					return nil, errors.Wrap(err, "DescribeQuery")
				}

				resultFile.Queries = append(resultFile.Queries, q)
			}
			fileSet = append(fileSet, resultFile)
		}

		result = append(result, fileSet)
	}

	return result, nil
}
