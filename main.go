package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	getter "github.com/hashicorp/go-getter"
)

var bundleURL = flag.String("bundle-url", "", "URL or path of the checks bundle to download")
var cacheDir = flag.String("cache-dir", "", "Trivy cache directory (default: auto-detected from trivy or $XDG_CACHE_HOME/trivy)")

func main() {
	flag.Parse()

	if err := run(context.Background()); err != nil {
		panic(err)
	}
}

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

func run(ctx context.Context) error {
	if *bundleURL == "" {
		return errors.New("bundle-url is required")
	}

	if *cacheDir == "" {
		dir, err := trivyCacheDir()
		if err != nil {
			return err
		}
		*cacheDir = dir
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "bundle-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	getters := map[string]getter.Getter{
		"file": &getter.FileGetter{Copy: true},
	}

	client := getter.Client{
		Ctx:             ctx,
		Src:             *bundleURL,
		Dst:             tmpDir,
		Getters:         getters,
		Mode:            getter.ClientModeDir,
		DisableSymlinks: true,
	}

	if err := client.Get(); err != nil {
		return err
	}

	contentDir := filepath.Join(*cacheDir, "policy", "content")
	if err := os.MkdirAll(contentDir, os.ModePerm); err != nil {
		return err
	}

	if err := os.CopyFS(contentDir, os.DirFS(tmpDir)); err != nil {
		return err
	}

	return nil
}
