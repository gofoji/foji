package output

import (
	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input"
	"github.com/sirupsen/logrus"
)

const (
	EmbedAll       = "EmbedAll"
	EmbedFileGroup = "EmbedFileGroup"
	EmbedFile      = "EmbedFile"
)

func HasEmbedOutput(o cfg.Output) bool {
	return hasAnyOutput(o, EmbedAll, EmbedFileGroup, EmbedFile)
}

func Embed(p cfg.Process, fn cfg.FileHandler, logger logrus.FieldLogger, groups []input.FileGroup, simulate bool) error {
	base := EmbedContext{
		Context: Context{Process:p, Logger: logger},
		FileGroups:     groups,
	}
	err := invokeProcess(p.Output[EmbedAll], p.RootDir, fn, logger, &base, simulate)
	if err != nil {
		return err
	}
	for _, ff := range groups {
		ctx := EmbedFileGroupContext{
			EmbedContext: base,
			FileGroup:    ff,
		}
		err := invokeProcess(p.Output[EmbedFileGroup], p.RootDir, fn, logger, &ctx, simulate)
		if err != nil {
			return err
		}

		for _, f := range ff.Files {
			ctx := EmbedFileContext{
				EmbedFileGroupContext: ctx,
				File:                  f,
			}
			err := invokeProcess(p.Output[EmbedFile], p.RootDir, fn, logger, &ctx, simulate)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type EmbedContext struct {
	Context
	FileGroups []input.FileGroup
}

type EmbedFileGroupContext struct {
	EmbedContext
	FileGroup input.FileGroup
}

type EmbedFileContext struct {
	EmbedFileGroupContext
	input.File
}
