package sql

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/rs/zerolog"

	"github.com/gofoji/foji/input"
	"github.com/gofoji/foji/runtime"
)

var (
	commentRE = regexp.MustCompile(`^--# `)
	paramRE   = regexp.MustCompile(`^@`)
)

type Repo interface {
	DescribeQuery(ctx context.Context, q *Query) error
}

type Parser struct {
	repo   Repo
	logger zerolog.Logger
	ctx    context.Context //nolint:containedctx
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
	return !runtime.IsGoType(r.Type) && !strings.Contains(r.Type, ".")
}

func (q Query) IsType(t ...string) bool {
	return slices.Contains(t, q.Type)
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
	} else {
		p.parseDescriptorHeader(l, q)
	}
}

const (
	positionName   = 0
	positionResult = 1
	positionType   = 2
)

func (p Parser) parseDescriptorHeader(l string, q *Query) {
	/*
		Example headers:
		--# FuncName
		--# FuncName ResultType
		--# FuncName ResultType QueryType
	*/
	ll := strings.Split(strings.TrimSpace(l), " ")
	if len(ll) > positionName {
		q.Name = ll[positionName]
	}

	if len(ll) > positionResult {
		q.Result.Type = ll[positionResult]
	}

	if len(ll) > positionType {
		q.Type = ll[positionType]
	}

	p.logger.Debug().Str("name", q.Name).Msg("Query Info")
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

		p.logger.Debug().Interface("param", p).Msg("Query Param")
	} else {
		param := Param{
			Ordinal: len(f.Params),
			Name:    pp[0],
			Type:    pp[1],
			Query:   f,
		}
		f.Params = append(f.Params, &param)

		p.logger.Debug().Interface("param", p).Msg("Query Param with Type")
	}
}

func Parse(ctx context.Context, logger zerolog.Logger, repo Repo, inGroups []input.FileGroup) (FileGroups, error) {
	result := make(FileGroups, len(inGroups))

	p := Parser{
		repo:   repo,
		logger: logger,
		ctx:    ctx,
	}

	for i, source := range inGroups {
		fileSet := make(FileGroup, len(source.Files))

		for j, f := range source.Files {
			logger.Debug().Str("filename", f.Name).Msg("SQL Parsing")

			resultFile, err := parseFile(ctx, f, logger, p)
			if err != nil {
				return nil, fmt.Errorf("parse:%q:%w", f.Name, err)
			}

			fileSet[j] = resultFile
		}

		result[i] = fileSet
	}

	return result, nil
}

func parseFile(ctx context.Context, f input.File, logger zerolog.Logger, p Parser) (File, error) {
	resultFile := File{File: f}

	statements := bytes.Split(f.Content, []byte(";"))
	for _, stmt := range statements {
		s := strings.TrimSpace(string(stmt))
		if s == "" {
			continue
		}

		logger.Trace().Str("sql", s).Msg("Parsing")

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

		err := p.repo.DescribeQuery(ctx, &q)
		if err != nil {
			return resultFile, fmt.Errorf("DescribeQuery: %w", err)
		}

		resultFile.Queries = append(resultFile.Queries, q)
	}

	return resultFile, nil
}
