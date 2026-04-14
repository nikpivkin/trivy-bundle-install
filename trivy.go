package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

var cacheDirRe = regexp.MustCompile(`--cache-dir\s+\S+\s+.*\(default "([^"]+)"\)`)

func trivyCacheDir() (string, error) {
	trivyPath, err := exec.LookPath("trivy")
	if err == nil {
		out, err := exec.Command(trivyPath, "--help").CombinedOutput()
		if err == nil {
			if m := cacheDirRe.FindSubmatch(out); m != nil {
				return string(m[1]), nil
			}
		}
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "trivy"), nil
}
