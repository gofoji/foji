package cmd

import "github.com/sirupsen/logrus"

func getLogger(quiet, verbose bool) logrus.FieldLogger {
	l := logrus.New()

	if quiet {
		l.SetLevel(logrus.FatalLevel)
	} else if verbose {
		l.SetLevel(logrus.TraceLevel)
	} else {
		l.SetLevel(logrus.InfoLevel)
	}
	return l
}
