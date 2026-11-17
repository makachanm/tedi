package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

type prompter struct {
	in  *bufio.Reader
	out io.Writer
}

func (p *prompter) printf(format string, a ...any) {
	fmt.Fprintf(p.out, format, a...)
}

func (p *prompter) ask(ctx context.Context, label string, def string) (string, error) {
	if def != "" {
		p.printf("%s [%s]: ", label, def)
	} else {
		p.printf("%s: ", label)
	}

	type answer struct {
		line string
		err  error
	}
	ch := make(chan answer, 1)
	go func() {
		line, err := p.in.ReadString('\n')
		ch <- answer{line, err}
	}()

	select {
	case <-ctx.Done():
		p.printf("\n")
		return "", ctx.Err()
	case a := <-ch:
		line := strings.TrimSpace(a.line)
		if a.err != nil && !(errors.Is(a.err, io.EOF) && line != "") {
			if errors.Is(a.err, io.EOF) {
				return "", errors.New("no input (stdin closed)")
			}
			return "", a.err
		}
		if line == "" {
			return def, nil
		}
		return line, nil
	}
}
