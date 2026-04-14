package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-units"
	getter "github.com/hashicorp/go-getter"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote"
)

type ociGetter struct {
	client *getter.Client
}

func (g *ociGetter) SetClient(c *getter.Client) { g.client = c }

func (g *ociGetter) ClientMode(_ *url.URL) (getter.ClientMode, error) {
	return getter.ClientModeDir, nil
}

func (g *ociGetter) GetFile(_ string, _ *url.URL) error {
	return errors.New("oci getter does not support single file downloads")
}

func (g *ociGetter) Get(dst string, u *url.URL) error {
	ctx := context.Background()
	if g.client != nil {
		ctx = g.client.Ctx
	}

	ref := u.Host + u.Path
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return err
	}

	_, manifestBytes, err := repo.FetchReference(ctx, repo.Reference.Reference)
	if err != nil {
		return err
	}
	defer manifestBytes.Close()

	var manifest ocispec.Manifest
	if err := json.NewDecoder(manifestBytes).Decode(&manifest); err != nil {
		return err
	}

	for _, layer := range manifest.Layers {
		slog.Debug("extracting layer", "digest", layer.Digest)
		rc, err := repo.Blobs().Fetch(ctx, layer)
		if err != nil {
			return err
		}
		var r io.ReadCloser = rc
		if g.client != nil && g.client.ProgressListener != nil {
			desc := fmt.Sprintf("%s (%s)", layer.Digest.Encoded()[:12], units.HumanSize(float64(layer.Size)))
			r = g.client.ProgressListener.TrackProgress(desc, 0, layer.Size, rc)
		}
		if err := extractTarGz(r, dst); err != nil {
			r.Close()
			return err
		}
		r.Close()
	}
	return nil
}

func extractTarGz(r io.Reader, dst string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dst, hdr.Name)
		if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid path: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}
