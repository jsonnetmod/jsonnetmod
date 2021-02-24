package jsonnet

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/google/go-jsonnet/formatter"
)

// FormatOpts
type FormatOpts struct {
	// ReplaceFile when enabled, will write formatted code to same file
	ReplaceFile bool `name:"write,w" usage:"write result to (source) file instead of stdout"`
	// PrintNames when enabled, will print formatted result of each file
	PrintNames bool `name:"list,l" `
}

// FormatFiles format jsonnet files
func FormatFiles(ctx context.Context, files []string, opt FormatOpts) error {
	log := logr.FromContextOrDiscard(ctx)

	writeFile := func(file string, data string) error {
		f, _ := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0)
		defer f.Close()
		_, err := io.WriteString(f, data)
		return err
	}

	for i := range files {
		file := files[i]

		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		original := string(data)

		formatted, err := Format(file, original)
		if err != nil {
			return err
		}

		if original != formatted {
			if opt.PrintNames {
				log.Info(fmt.Sprintf("`%s` formatted.", file))
			}

			if opt.ReplaceFile {
				if err := writeFile(file, formatted); err != nil {
					return err
				}
			} else {
				fmt.Printf(`
// %s 

%s
`, file, formatted)
			}
		} else {
			if opt.PrintNames {
				log.Info(fmt.Sprintf("`%s` no changes.", file))
			}
		}
	}

	return nil

}

func Format(filename string, content string) (string, error) {
	return formatter.Format(filename, content, formatter.DefaultOptions())
}
