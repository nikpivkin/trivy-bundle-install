package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	getter "github.com/hashicorp/go-getter"
)

var bundleURL = flag.String("bundle-url", "", "URL or path of the checks bundle to download")
var cacheDir = flag.String("cache-dir", "", "Trivy cache directory (default: auto-detected from trivy or $XDG_CACHE_HOME/trivy)")

func main() {
	flag.Parse()

	if err := run(context.Background()); err != nil {
		slog.Error("failed", "err", err)
		os.Exit(1)
	}
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
		slog.Debug("auto-detected cache dir", "path", *cacheDir)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "bundle-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	slog.Info("downloading bundle", "url", *bundleURL)

	getters := map[string]getter.Getter{
		"file":  &getter.FileGetter{Copy: true},
		"http":  &getter.HttpGetter{},
		"https": &getter.HttpGetter{},
		"oci":   &ociGetter{},
	}

	client := getter.Client{
		Ctx:              ctx,
		Src:              *bundleURL,
		Dst:              tmpDir,
		Getters:          getters,
		Mode:             getter.ClientModeDir,
		ProgressListener: &progressTracker{},
		DisableSymlinks:  true,
	}

	if err := client.Get(); err != nil {
		return err
	}

	contentDir := filepath.Join(*cacheDir, "policy", "content")
	slog.Info("installing bundle", "path", contentDir)

	if err := os.RemoveAll(contentDir); err != nil {
		return err
	}
	if err := os.MkdirAll(contentDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.CopyFS(contentDir, os.DirFS(tmpDir)); err != nil {
		return err
	}

	slog.Info("bundle installed successfully")
	return nil
}
