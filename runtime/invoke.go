package runtime

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func Invoke(file string, commandLine []string) error {
	if len(commandLine) < 1 {
		return nil
	}
	name := commandLine[0]
	args := make([]string, len(commandLine)-1)
	for x, s := range commandLine[1:] {
		args[x] = os.Expand(s, func(s string) string {
			if s == "FILE" {
				return file
			}
			return ""
		})
	}
	var cmd *exec.Cmd
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	cmd = exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "error invoking post process command '%s %s'", name, strings.Join(args, " "))
	}
	return nil
}
