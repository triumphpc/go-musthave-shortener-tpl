// Static analytic service. Include static analytic packages:
// - golang.org/x/tools/go/analysis/passes
// - all SA classes staticcheck.io
// - Go-critic and nilerr linters
// - OsExitAnalyzer to check os.Exit calls in main packages
//
// How to run:
//
// ./cmd/staticlint/main ./...

package main

import (
	"encoding/json"
	"fmt"
	goc "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/gostaticanalysis/nilerr"
	"github.com/triumphpc/go-musthave-shortener-tpl/pkg/myanalyzer"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"os"
	"path/filepath"
	"strings"
)

// Config  name config
const Config = `config/config.json`

// ConfigData describe config file
type ConfigData struct {
	Staticcheck []string
	Stylecheck  []string
}

func main() {
	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		panic(err)
	}

	var cfg ConfigData
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	mychecks := []*analysis.Analyzer{
		myanalyzer.OsExitAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		goc.Analyzer,    // Go-critic
		nilerr.Analyzer, // nilerr
	}

	// Enable all SA checks
	for _, v := range staticcheck.Analyzers {
		for _, sc := range cfg.Staticcheck {
			if strings.HasPrefix(v.Analyzer.Name, sc) {
				mychecks = append(mychecks, v.Analyzer)
			}
		}
	}

	for _, v := range stylecheck.Analyzers {
		for _, sc := range cfg.Stylecheck {
			if strings.HasPrefix(v.Analyzer.Name, sc) {
				mychecks = append(mychecks, v.Analyzer)
			}
		}
	}

	fmt.Println("Run static checks:\n", mychecks)

	multichecker.Main(
		mychecks...,
	)
}
