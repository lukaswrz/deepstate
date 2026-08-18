package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func findStorePath() (string, error) {
	cmd := exec.Command("nix", "eval", "--raw", "--expr", storeDirExpr)
	p, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not Nix eval %s: %w", storeDirExpr, err)
	}
	return string(p), nil
}

func matchesExclude(exclude, path string) bool {
	var err error

	exclude, err = filepath.Abs(exclude)
	if err != nil {
		return false
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(exclude, path)
	if err != nil {
		return false
	}

	return filepath.IsLocal(rel)
}

func matchesExcludes(excludes []string, path string) bool {
	for _, exclude := range excludes {
		if matchesExclude(exclude, path) {
			return true
		}
	}

	return false
}
