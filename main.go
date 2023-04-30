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

var (
	selectAll          bool
	selectAllNoConfirm bool
	dryRun             bool
)

func init() {
	flag.BoolVar(&selectAll, "a", false, "selects everything by default")
	flag.BoolVar(&selectAllNoConfirm, "A", false, "selects and updates everything by default without UI")
	flag.BoolVar(&dryRun, "d", false, "does not run go get if true")
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
	if len(args) == 0 {
		return updateGoMod("go.mod")
	}

	var errs []error
	for _, arg := range args {
		err := updateGoMod(arg)
		if err != nil {
			err = fmt.Errorf("could not update %q: %w", arg, err)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func updateGoMod(path string) error {
	modules, err := parseDirectDeps(path)
	if err != nil {
		return fmt.Errorf("parse direct deps: %w", err)
	}

	if len(modules) == 0 {
		return errors.New("no direct dependencies found")
	}

	if selectAllNoConfirm {
		return runGoGet(path, modulesToPaths(modules))
	}

	selected, err := runUI(modules)
	if err != nil {
		return fmt.Errorf("ui failed: %w", err)
	}

	if len(selected) == 0 {
		return errors.New("no modules selected")
	}

	return runGoGet(path, modulesToPaths(selected))
}

func parseDirectDeps(path string) ([]module.Version, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ast, err := modfile.Parse(path, content, nil)
	if err != nil {
		return nil, fmt.Errorf("parse go.mod: %w", err)
	}

	var modules []module.Version
	for _, req := range ast.Require {
		if req.Indirect {
			continue
		}

		modules = append(modules, req.Mod)
	}

	return modules, nil
}

func runUI(modules []module.Version) ([]module.Version, error) {
	f, err := fzf.New(fzf.WithNoLimit(true))
	if err != nil {
		return nil, fmt.Errorf("create fzf: %w", err)
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
		return nil, fmt.Errorf("fuzzy select: %w", err)
	}

	selected := make([]module.Version, len(indices))
	for i, idx := range indices {
		selected[i] = modules[idx]
	}

	return selected, nil
}

func runGoGet(path string, selected []string) error {
	fmt.Println("go get", strings.Join(selected, " "))

	if dryRun {
		return nil
	}

	workDir := filepath.Dir(path)
	cmd := exec.Command("go", append([]string{"get"}, selected...)...)
	cmd.Dir = workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("run go get: %w", err)
	}

	return err
}

func modulesToPaths(modules []module.Version) []string {
	paths := make([]string, len(modules))
	for i := range modules {
		paths[i] = modules[i].Path
	}
	return paths
}
