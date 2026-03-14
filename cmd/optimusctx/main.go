package main

import (
	"os"

	"github.com/niccrow/optimusctx/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
