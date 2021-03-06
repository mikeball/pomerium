package cluster

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pomerium/pomerium/internal/log"
)

type cmdOption func(*exec.Cmd)

func withArgs(args ...string) cmdOption {
	return func(cmd *exec.Cmd) {
		cmd.Args = append([]string{"kubectl"}, args...)
	}
}

func withEnv(env ...string) cmdOption {
	return func(cmd *exec.Cmd) {
		cmd.Env = append(os.Environ(), env...)
	}
}

func withStdin(rdr io.Reader) cmdOption {
	return func(cmd *exec.Cmd) {
		cmd.Stdin = rdr
	}
}

func withStdout(w io.Writer) cmdOption {
	return func(cmd *exec.Cmd) {
		cmd.Stdout = w
	}
}

func withWorkingDir(wd string) cmdOption {
	return func(cmd *exec.Cmd) {
		cmd.Dir = wd
	}
}

func run(ctx context.Context, name string, options ...cmdOption) error {
	cmd := commandContext(ctx, name)
	for _, o := range options {
		o(cmd)
	}
	if cmd.Stderr == nil {
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to create stderr pipe for %s: %w", name, err)
		}
		go cmdLogger(ctx, stderr)
		defer stderr.Close()
	}
	if cmd.Stdout == nil {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdout pipe for %s: %w", name, err)
		}
		go cmdLogger(ctx, stdout)
		defer stdout.Close()
	}

	log.Debug(ctx).Strs("args", cmd.Args).Msgf("running %s", name)
	return cmd.Run()
}

func cmdLogger(ctx context.Context, rdr io.Reader) {
	s := bufio.NewScanner(rdr)
	for s.Scan() {
		log.Debug(ctx).Msg(s.Text())
	}
}
