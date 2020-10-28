package welder

import (
	"context"
	"fmt"
	"time"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input"
	"github.com/gofoji/foji/input/db"
	"github.com/gofoji/foji/input/db/pg"
	"github.com/gofoji/foji/input/openapi"
	"github.com/gofoji/foji/input/proto"
	"github.com/gofoji/foji/input/sql"
	sqlDB "github.com/gofoji/foji/input/sql/pg"
	"github.com/gofoji/foji/log"
	"github.com/gofoji/foji/output"
	"github.com/gofoji/foji/runtime"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type InputFiles struct {
	Config cfg.FileInput
	Loaded input.FileGroup
}

type Welder struct {
	logger  logrus.FieldLogger
	config  cfg.Config
	targets []cfg.Process

	ctx       context.Context
	conn      *pgx.Conn
	resources map[string]InputFiles
}

type Processor struct {
	guard func(o cfg.Output) bool
	run   func(simulate bool, p cfg.Process, ff []input.FileGroup) error
}

// New creates a new welder.
func New(logger logrus.FieldLogger, config cfg.Config, targets []cfg.Process) *Welder {
	w := Welder{
		ctx:       context.Background(),
		logger:    logger,
		config:    config,
		targets:   targets,
		resources: map[string]InputFiles{},
		conn:      nil,
	}

	return &w
}

func (w *Welder) Run(simulate bool) error {
	pp := []Processor{
		{guard: output.HasDBOutput, run: w.dbProcess},
		{guard: output.HasSQLOutput, run: w.sqlProcess},
		{guard: output.HasEmbedOutput, run: w.embedProcess},
		{guard: output.HasProtoOutput, run: w.protoProcess},
		{guard: output.HasOpenAPIOutput, run: w.apiProcess},
	}

	for _, p := range w.targets {
		w.logger.WithField("Process", p.ID).Info("Welding")

		ff, err := w.getProcessFiles(p)
		if err != nil {
			return err
		}

		for _, processor := range pp {
			if !processor.guard(p.Output) {
				continue
			}

			start := time.Now()

			err = processor.run(simulate, p, ff)
			if err != nil {
				return err
			}

			w.logger.WithField("Process", p.ID).WithField("duration", time.Since(start)).Trace("Welding")
		}
	}

	return nil
}

func (w *Welder) embedProcess(simulate bool, p cfg.Process, ff []input.FileGroup) error {
	w.logger.Info("Embed")

	return output.Embed(p, w.postProcessor(p), w.logger, ff, simulate)
}

func (w *Welder) apiProcess(simulate bool, p cfg.Process, ff []input.FileGroup) error {
	w.logger.Info("OpenAPI")

	pp, err := openapi.Parse(w.ctx, w.logger, ff)
	if err != nil {
		return fmt.Errorf("openapi.Parse:%w", err)
	}

	return output.OpenAPI(p, w.postProcessor(p), w.logger, pp, simulate)
}

func (w *Welder) protoProcess(simulate bool, p cfg.Process, ff []input.FileGroup) error {
	w.logger.Info("Proto")

	pp, err := proto.Parse(w.ctx, w.logger, ff)
	if err != nil {
		return fmt.Errorf("proto.Parse:%w", err)
	}

	return output.Proto(p, w.postProcessor(p), w.logger, pp, simulate)
}

func (w *Welder) sqlProcess(simulate bool, p cfg.Process, ff []input.FileGroup) error {
	w.logger.Info("SQL")

	err := w.initDBConnection()
	if err != nil {
		return err
	}

	sqlFiles, err := sql.Parse(w.ctx, w.logger, sqlDB.New(w.conn), ff)
	if err != nil {
		return fmt.Errorf("sql.Parse:%w", err)
	}

	return output.SQL(p, w.postProcessor(p), w.logger, sqlFiles, simulate)
}

func (w *Welder) dbProcess(simulate bool, p cfg.Process, _ []input.FileGroup) error {
	w.logger.Info("DB")

	err := w.initDBConnection()
	if err != nil {
		return err
	}

	s, err := w.parseDB()
	if err != nil {
		return err
	}

	return output.DB(p, w.postProcessor(p), w.logger, s, simulate)
}

func (w *Welder) getResource(resource string) (input.FileGroup, error) {
	r, ok := w.resources[resource]
	if !ok {
		w.logger.WithField("resource", resource).Trace("getResource")

		in, ok := w.config.Files[resource]
		if !ok {
			return r.Loaded, errors.Errorf("invalid resource reference:%s", resource)
		}

		f, err := input.Parse(w.ctx, w.logger, in)
		if err != nil {
			return r.Loaded, fmt.Errorf("input.Parse:%w", err)
		}

		r = InputFiles{
			Config: in,
			Loaded: f,
		}

		w.resources[resource] = r
	}

	return r.Loaded, nil
}

func (w *Welder) getProcessFiles(p cfg.Process) ([]input.FileGroup, error) {
	result := make([]input.FileGroup, len(p.Resources))

	for i, resource := range p.Resources {
		f, err := w.getResource(resource)
		if err != nil {
			return nil, err
		}

		result[i] = f
	}

	if !p.Files.IsEmpty() {
		f, err := input.Parse(w.ctx, w.logger, p.Files)
		if err != nil {
			return nil, fmt.Errorf("input.Parse:%w", err)
		}

		result = append(result, f)
	}

	return result, nil
}

func (w *Welder) parseDB() (db.DB, error) {
	if w.conn == nil {
		return nil, errors.New("DB Not Initialized")
	}

	repo := pg.New(w.conn, w.logger)

	s, err := repo.Get(w.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "parsing DB schema")
	}

	return s.Filter(w.config.DB.Filter), nil
}

func (w *Welder) initDBConnection() error {
	if w.conn != nil {
		return nil
	}

	if w.config.DB.Connection == "" {
		return errors.New("missing DB connection")
	}

	w.logger.WithField("Connection", w.config.DB.Connection).Debug("Loading Database")

	config, err := pgx.ParseConfig(w.config.DB.Connection)
	if err != nil {
		w.logger.WithError(err).Fatal("pgx.ParseConfig")
	}

	config.Logger = log.NewLogger(w.logger)

	conn, err := pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		w.logger.WithError(err).Fatal("Unable to connect to database")
	}

	dt, err := pgxtype.LoadDataType(context.Background(), conn, conn.ConnInfo(), "_name")
	if err != nil {
		w.logger.WithError(err).Fatal("Unable to Load Data Types")
	}

	conn.ConnInfo().RegisterDataType(dt)

	w.conn = conn

	return nil
}

func (w Welder) postProcessor(p cfg.Process) cfg.FileHandler {
	if len(p.Post) == 0 || len(p.Post[0]) == 0 {
		return nil
	}

	return func(filename string) error {
		for _, post := range p.Post {
			w.logger.WithField("cmd", post.Join(" ")).Debug("post processor")

			start := time.Now()

			err := runtime.Invoke(filename, post)
			if err != nil {
				return fmt.Errorf("runtime.Invoke:%w", err)
			}

			w.logger.WithField("duration", time.Since(start)).Trace("post processor")
		}

		return nil
	}
}

func (w *Welder) Close() {
	if w.conn != nil {
		err := w.conn.Close(w.ctx)
		if err != nil {
			w.logger.WithError(err).Error("DB Connection Close")
		}
	}
}
