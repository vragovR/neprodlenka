package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
)

var args struct {
	Run   *RunCmd   `arg:"subcommand:run" help:"запустить бота и HTTP-сервер"`
	Admin *AdminCmd `arg:"subcommand:admin" help:"управление администраторами"`
}

type command interface {
	Execute() error
}

func main() {
	p := arg.MustParse(&args)

	cmd, ok := p.Subcommand().(command)
	if !ok {
		_ = p.FailSubcommand("missing subcommand", p.SubcommandNames()...)
	}
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
