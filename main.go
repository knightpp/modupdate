package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/koki-develop/go-fzf"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

var selectAll bool

func init() {
	flag.BoolVar(&selectAll, "a", false, "selects everything by default")
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func run() error {
	args := flag.Args()
	gomodPath := "go.mod"
	if len(args) != 0 {
		gomodPath = args[0]
	}

	content, err := os.ReadFile(gomodPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	ast, err := modfile.Parse(gomodPath, content, nil)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	var modules []module.Version
	for _, req := range ast.Require {
		if req.Indirect {
			continue
		}

		modules = append(modules, req.Mod)
	}

	f, err := fzf.New(fzf.WithNoLimit(true))
	if err != nil {
		return err
	}

	var findOpts []fzf.FindOption
	if selectAll {
		findOpts = append(findOpts, fzf.WithDefaultSelectionAll())
	}

	indices, err := f.Find(modules, func(i int) string {
		mod := modules[i]
		return mod.Path + " " + mod.Version
	}, findOpts...)
	if err != nil {
		return fmt.Errorf("fuzzy select: %w", err)
	}

	if len(indices) == 0 {
		return errors.New("no modules selected")
	}

	selected := make([]string, len(indices))
	for i, idx := range indices {
		selected[i] = modules[idx].Path
	}

	fmt.Println("go get", strings.Join(selected, " "))

	workDir := filepath.Dir(gomodPath)
	cmd := exec.Command("go", append([]string{"get"}, selected...)...)
	cmd.Dir = workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("run go get: %w", err)
	}

	return nil
}
