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
	version            bool
)

func init() {
	flag.BoolVar(&selectAll, "a", false, "select everything by default")
	flag.BoolVar(&selectAllNoConfirm, "A", false, "select and update everything without confirmation")
	flag.BoolVar(&dryRun, "d", false, "dry run, just print what will be executed")
	flag.BoolVar(&version, "v", false, "show version information")
}

func main() {
	flag.Parse()

	if version {
		fmt.Printf("revision:\t%s\ndate:\t\t%s\ndirty:\t\t%v\ncompiler:\t%s\n", vcsCommit, vcsTime, vcsModified, compiler)
		return
	}

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

func updateGoMod(gomodPath string) error {
	fi, err := os.Stat(gomodPath)
	if err != nil {
		return fmt.Errorf("stat %q: %w", gomodPath, err)
	}

	if fi.IsDir() {
		gomodPath = filepath.Join(gomodPath, "go.mod")
	}

	modules, err := parseDirectDeps(gomodPath)
	if err != nil {
		return fmt.Errorf("parse direct deps: %w", err)
	}

	if len(modules) == 0 {
		return errors.New("no direct dependencies found")
	}

	if selectAllNoConfirm {
		return runGoGet(gomodPath, modulesToPaths(modules))
	}

	selected, err := runUI(modules)
	if err != nil {
		return fmt.Errorf("ui failed: %w", err)
	}

	if len(selected) == 0 {
		return errors.New("no modules selected")
	}

	return runGoGet(gomodPath, modulesToPaths(selected))
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
		findOpts = append(findOpts, fzf.WithPreselectAll(true))
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
