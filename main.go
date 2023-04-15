package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]
	filePath := "go.mod"
	if len(args) != 0 {
		filePath = args[0]
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	ast, err := modfile.Parse(filePath, content, nil)
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

	indices, err := fuzzyfinder.FindMulti(
		modules,
		func(i int) string {
			mod := modules[i]
			return mod.Path + " " + mod.Version
		},
		fuzzyfinder.WithHeader("Select modules to update"),
	)
	if err != nil {
		return fmt.Errorf("fuzzy select: %w", err)
	}

	selected := make([]string, len(indices))
	for i, idx := range indices {
		selected[i] = modules[idx].Path
	}

	fmt.Println("go get", strings.Join(selected, " "))

	cmd := exec.Command("go", append([]string{"get"}, selected...)...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("run go get: %w", err)
	}

	return nil
}
