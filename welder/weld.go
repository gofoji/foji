package welder

import (
	"context"

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
	targets cfg.Processes // Final list of processes

	ctx       context.Context
	conn      *pgx.Conn
	resources map[string]InputFiles
}

// New creates a new welder
func New(logger logrus.FieldLogger, config cfg.Config, targets cfg.Processes) *Welder {
	w := Welder{ctx: context.Background(), logger: logger, config: config, targets: targets, resources: map[string]InputFiles{}}
	return &w
}

func (w *Welder) Run(simulate bool) error {
	for _, p := range w.targets {
		w.logger.WithField("Process", p.ID).Info("Welding")

		ff, err := w.getProcessFiles(p)
		if err != nil {
			return err
		}

		if output.HasDBOutput(p.Output) {
			w.logger.Info("DB")

			err = w.initDBConnection()
			if err != nil {
				return err
			}

			s, err := w.parseDB()
			if err != nil {
				return err
			}

			err = output.DB(p, w.postProcessor(p), w.logger, s, simulate)
			if err != nil {
				return err
			}
		}

		if output.HasSQLOutput(p.Output) {
			w.logger.Info("SQL")
			err = w.initDBConnection()
			if err != nil {
				return err
			}

			sqlFiles, err := sql.Parse(w.ctx, w.logger, sqlDB.New(w.conn), ff)
			if err != nil {
				return err
			}

			err = output.SQL(p, w.postProcessor(p), w.logger, sqlFiles, simulate)
			if err != nil {
				return err
			}
		}

		if output.HasEmbedOutput(p.Output) {
			w.logger.Info("Embed")

			err = output.Embed(p, w.postProcessor(p), w.logger, ff, simulate)
			if err != nil {
				return err
			}
		}

		if output.HasProtoOutput(p.Output) {
			w.logger.Info("Proto")

			pp, err := proto.Parse(w.ctx, w.logger, ff)
			if err != nil {
				return err
			}

			err = output.Proto(p, w.postProcessor(p), w.logger, pp,simulate)
			if err != nil {
				return err
			}
		}

		if output.HasOpenAPIOutput(p.Output) {
			w.logger.Info("OpenAPI")

			pp, err := openapi.Parse(w.ctx, w.logger, ff)
			if err != nil {
				return err
			}

			err = output.OpenAPI(p, w.postProcessor(p), w.logger, pp, simulate)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Welder) getResource(resource string) (input.FileGroup, error) {
	r, ok := w.resources[resource]
	if !ok {
		in, ok := w.config.Files[resource]
		if !ok {
			return r.Loaded, errors.Errorf("invalid resource reference:%s", resource)
		}

		f, err := input.Parse(w.ctx, w.logger, in)
		if err != nil {
			return r.Loaded, err
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
	var result []input.FileGroup
	for _, resource := range p.Resources {
		f, err := w.getResource(resource)
		if err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	if !p.Files.IsEmpty() {
		f, err := input.Parse(w.ctx, w.logger, p.Files)
		if err != nil {
			return nil, err
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
	if len(p.Post) == 0 {
		return nil
	}

	return func(filename string) error {
		for _, post := range p.Post {
			w.logger.WithField("cmd", post).Debug("post processor")
			err := runtime.Invoke(filename, post)
			if err != nil {
				return err
			}
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
