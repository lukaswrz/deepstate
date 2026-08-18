package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
	"hack.moontide.ink/lukas/binfo"
)

var bi = binfo.MustGet()

const storeDirExpr = "builtins.storeDir"

func main() {
	var (
		verbose          bool
		store            string
		excludes         []string
		defaultExcludes  []string
		noDefaultExclude bool
	)

	cli.VersionPrinter = func(cmd *cli.Command) {
		_, _ = fmt.Fprintf(
			cmd.Root().Writer,
			"%s\n",
			bi.Summarize(
				cmd.Name,
				cmd.Version,
				binfo.Multiline|binfo.Build|binfo.VCS|binfo.Module|binfo.CGO,
			),
		)
	}

	app := &cli.Command{
		Name:        "deepstate",
		Version:     bi.Module.Version,
		Description: "Scan for leftovers",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "verbose output",
				Destination: &verbose,
			},
			&cli.StringFlag{
				Name:        "store",
				Usage:       "path to the Nix store",
				Destination: &store,
				DefaultText: "path gathered via " + storeDirExpr,
			},
			&cli.StringSliceFlag{
				Name:        "exclude",
				Usage:       "exclude this file or files in this directory",
				Destination: &excludes,
			},
			&cli.StringSliceFlag{
				Name:        "default-exclude",
				Usage:       "default exclude this file or files in this directory",
				Destination: &defaultExcludes,
			},
			&cli.BoolFlag{
				Name:        "no-default-exclude",
				Usage:       "disable usage of default excludes",
				Destination: &noDefaultExclude,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			roots := cmd.Args().Slice()
			if len(roots) == 0 {
				roots = append(roots, ".")
			}

			if store == "" {
				var err error
				store, err = findStorePath()
				if err != nil {
					return err
				}
			}

			if !noDefaultExclude {
				excludes = slices.Concat(defaultExcludes, excludes)
			}

			return scan(cmd.Writer, cmd.ErrWriter, roots, store, excludes, verbose)
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s\n", err)
		os.Exit(1)
	}
}

func scan(w, ew io.Writer, roots []string, store string, excludes []string, verbose bool) error {
	if verbose {
		fmt.Fprintf(ew, "using store %s\n", store)
		if len(excludes) > 0 {
			fmt.Fprintf(ew, "excluding %s\n", strings.Join(excludes, ", "))
		}
	}

	didErr := false
	callback := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			didErr = true
			fmt.Fprintf(ew, "walk: %s\n", err)
			return nil
		}

		err = handle(w, path, d, store, excludes)
		if err == filepath.SkipDir || err == filepath.SkipAll {
			return err
		}
		if err != nil {
			didErr = true
			fmt.Fprintln(ew, err)
		}

		return nil
	}

	for _, root := range roots {
		if verbose {
			fmt.Fprintf(ew, "scanning %s\n", root)
		}
		if err := filepath.WalkDir(root, callback); err != nil {
			return err
		}
	}

	if didErr {
		return errors.New("errors occurred while walking tree")
	}

	return nil
}

func handle(w io.Writer, path string, d fs.DirEntry, store string, excludes []string) error {
	if matchesExcludes(slices.Concat(excludes, []string{store}), path) {
		if d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	if d.IsDir() {
		return nil
	}

	target, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("cannot resolve: %w", err)
	}

	if matchesExclude(store, target) {
		return nil
	}

	fmt.Fprintln(w, path)

	return nil
}
