package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const Timeout = time.Second * 15

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

			return os.Getenv(s)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%w: error invoking post process command '%s %s'", err, name, strings.Join(args, " "))
	}

	return nil
}
