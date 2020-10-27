package cmd

import "github.com/sirupsen/logrus"

func getLogger(quiet, trace, verbose bool) logrus.FieldLogger {
	l := logrus.New()

	switch {
	case quiet:
		l.SetLevel(logrus.FatalLevel)
	case trace:
		l.SetLevel(logrus.TraceLevel)
	case verbose:
		l.SetLevel(logrus.DebugLevel)
	default:
		l.SetLevel(logrus.InfoLevel)
	}

	return l
}
