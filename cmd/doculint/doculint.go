package main

import (
	"github.com/george-e-shaw-iv/doculint/internal/doculint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(&doculint.Analyzer)
}
