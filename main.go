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
	fSelectAll          bool
	fSelectAllNoConfirm bool
	fHidden             bool
	fDryRun             bool
	fVersion            bool
	fSort               bool
)

func init() {
	flag.BoolVar(&fSelectAll, "a", false, "select everything by default")
	flag.BoolVar(&fSelectAllNoConfirm, "A", false, "select and update everything without confirmation")
	flag.BoolVar(&fDryRun, "d", false, "dry run, just print what will be executed")
	flag.BoolVar(&fVersion, "v", false, "show version information")
	flag.BoolVar(&fSort, "s", false, "sort require lines")
	flag.BoolVar(&fHidden, "h", false, "show indirect")
}

func main() {
	flag.Parse()

	if fVersion {
		if version != "" {
			fmt.Printf("version:\t%s\n", version)
		}
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

	gomod, err := parseGoMod(gomodPath)
	if err != nil {
		return fmt.Errorf("parse direct deps: %w", err)
	}

	modules := extractDeps(gomod)

	if len(modules) == 0 {
		return errors.New("no direct dependencies found")
	}

	if fSelectAllNoConfirm {
		return runGoGet(gomodPath, modulesToPaths(modules))
	}

	if fSort {
		return sortImports(gomodPath, gomod)
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

func parseGoMod(path string) (*modfile.File, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ast, err := modfile.Parse(path, content, nil)
	if err != nil {
		return nil, fmt.Errorf("parse go.mod: %w", err)
	}

	return ast, nil
}

func extractDeps(gomod *modfile.File) []module.Version {
	var modules []module.Version
	for _, req := range gomod.Require {
		if !fHidden && req.Indirect {
			continue
		}

		modules = append(modules, req.Mod)
	}

	return modules
}

func runUI(modules []module.Version) ([]module.Version, error) {
	f, err := fzf.New(fzf.WithNoLimit(true))
	if err != nil {
		return nil, fmt.Errorf("create fzf: %w", err)
	}

	var findOpts []fzf.FindOption
	if fSelectAll {
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

	if fDryRun {
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

func sortImports(gomodPath string, gomod *modfile.File) error {
	var direct, indirect []modfile.Require
	for _, req := range gomod.Require {
		if req.Indirect {
			indirect = append(indirect, *req)
		} else {
			direct = append(direct, *req)
		}
		gomod.DropRequire(req.Mod.Path)
	}

	for _, dep := range direct {
		gomod.AddNewRequire(dep.Mod.Path, dep.Mod.Version, false)
	}
	for _, dep := range indirect {
		gomod.AddNewRequire(dep.Mod.Path, dep.Mod.Version, true)
	}

	gomod.Cleanup()

	gomod.SetRequireSeparateIndirect(gomod.Require)

	bytes, err := gomod.Format()
	if err != nil {
		return fmt.Errorf("format gomod: %w", err)
	}

	info, err := os.Stat(gomodPath)
	if err != nil {
		return fmt.Errorf("stat gomod: %w", err)
	}

	err = os.WriteFile(gomodPath, bytes, info.Mode())
	if err != nil {
		return fmt.Errorf("write gomod: %w", err)
	}

	return nil
}
