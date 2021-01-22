package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func Pageln(i ...interface{}) error {
	return FPageln(strings.NewReader(fmt.Sprint(i...)))
}

// FPageln invokes the systems pager with the supplied data.
// If the PAGER environment variable is empty, no pager is used.
// If the PAGER environment variable is unset, use GNU less with convenience flags.
func FPageln(r io.Reader) error {
	pager, ok := os.LookupEnv("PAGER")

	if !ok {
		// --RAW-CONTROL-CHARS  Honors colors from diff. Must be in all caps, otherwise display issues occur.
		// --quit-if-one-screen Closer to the git experience.
		// --no-init            Don't clear the screen when exiting.
		pager = "less --RAW-CONTROL-CHARS --quit-if-one-screen --no-init"
	}

	sh, err := syntax.NewParser().Parse(strings.NewReader(pager), "")
	if err != nil {
		return err
	}

	runner, err := interp.New(interp.StdIO(r, os.Stdout, os.Stderr))
	if err != nil {
		return err
	}

	return runner.Run(context.Background(), sh)
}
