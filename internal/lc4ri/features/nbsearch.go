package features

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const nbsearchName = "nbsearch"

//go:embed images/nbsearch-solr/**
var nbsearchSolrFiles embed.FS

type NBSearch struct {
	RuntimeDir string
}

func (f NBSearch) Enable() error {
	composeFile := f.composeFile()
	if _, err := os.Stat(composeFile); err == nil {
		return fmt.Errorf("nbsearch is already enabled: %s", composeFile)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat nbsearch compose file %s: %w", composeFile, err)
	}

	if err := f.writeNotebookBuildScript(); err != nil {
		return err
	}
	if err := f.writeSolrFiles(); err != nil {
		return err
	}

	if err := os.WriteFile(composeFile, []byte(nbsearchComposeTemplate()), 0o644); err != nil {
		return fmt.Errorf("write nbsearch compose file %s: %w", composeFile, err)
	}

	return nil
}

func (f NBSearch) Disable() error {
	composeFile := f.composeFile()
	if _, err := os.Stat(composeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("nbsearch is not enabled: %s", composeFile)
		}
		return fmt.Errorf("stat nbsearch compose file %s: %w", composeFile, err)
	}

	if err := os.Remove(composeFile); err != nil {
		return fmt.Errorf("remove nbsearch compose file %s: %w", composeFile, err)
	}
	if err := removePath(filepath.Join(f.RuntimeDir, "notebook", nbsearchName)); err != nil {
		return err
	}
	if err := removePath(filepath.Join(f.RuntimeDir, "images", "nbsearch-solr")); err != nil {
		return err
	}

	return nil
}

func (f NBSearch) composeFile() string {
	return filepath.Join(f.RuntimeDir, "compose."+nbsearchName+".yaml")
}

func (f NBSearch) writeNotebookBuildScript() error {
	dir := filepath.Join(f.RuntimeDir, "notebook", nbsearchName)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("nbsearch notebook build directory already exists: %s", dir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat nbsearch notebook build directory %s: %w", dir, err)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create nbsearch notebook build directory %s: %w", dir, err)
	}
	if err := writeFile(filepath.Join(dir, "build.sh"), nbsearchNotebookBuildScript(), 0o644); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, "initial-index-hook.sh"), nbsearchInitialIndexHookScript(), 0o644); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, "initial-index.sh"), nbsearchInitialIndexScript(), 0o644); err != nil {
		return err
	}
	return nil
}

func (f NBSearch) writeSolrFiles() error {
	sourceRoot := "images/nbsearch-solr"
	targetRoot := filepath.Join(f.RuntimeDir, "images", "nbsearch-solr")

	return fs.WalkDir(nbsearchSolrFiles, sourceRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath := strings.TrimPrefix(path, sourceRoot)
		relativePath = strings.TrimPrefix(relativePath, "/")
		targetPath := filepath.Join(targetRoot, relativePath)
		if entry.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create nbsearch asset directory %s: %w", targetPath, err)
			}
			return nil
		}

		data, err := nbsearchSolrFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded nbsearch asset %s: %w", path, err)
		}
		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
			return fmt.Errorf("write nbsearch asset %s: %w", targetPath, err)
		}
		return nil
	})
}

func removePath(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}

func writeFile(path, content string, perm fs.FileMode) error {
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func nbsearchNotebookBuildScript() string {
	return `set -eu
script_dir=$(dirname "$0")
jupyter labextension enable nbsearch --level=system
install -m 0755 "${script_dir}/initial-index-hook.sh" /usr/local/bin/before-notebook.d/nbsearch-initial-index.sh
install -m 0755 "${script_dir}/initial-index.sh" /usr/local/bin/lc4ri-nbsearch-initial-index
fix-permissions "${CONDA_DIR}"
fix-permissions "/home/${NB_USER}"
`
}

func nbsearchInitialIndexHookScript() string {
	return `#!/bin/bash

log=/tmp/lc4ri-nbsearch-initial-index.log
/usr/local/bin/lc4ri-nbsearch-initial-index >> "${log}" 2>&1 &
`
}

func nbsearchInitialIndexScript() string {
	return `#!/bin/bash

set -xe

if [[ -f "/home/${NB_USER}/.nbsearch/config_local.py" ]]; then
  exit 0
fi

while ! nc -z nbsearch-solr 8983; do
  sleep 0.5
done

while ! nc -z nbsearch-minio 9000; do
  sleep 0.5
done

while ! curl http://nbsearch-solr:8983/solr/jupyter-cell/admin/ping | grep '"status":"OK"'; do
  sleep 0.5
done

jupyter nbsearch update-index --debug "${CONDA_DIR}/etc/jupyter/jupyter_notebook_config.py" local
`
}

func nbsearchComposeTemplate() string {
	return `volumes:
  nbsearch_solr_data:
  nbsearch_minio_data:

services:
  notebook:
    depends_on:
      - nbsearch-solr
      - nbsearch-minio
    environment:
      NBSEARCHDB_SOLR_BASE_URL: http://nbsearch-solr:8983
      NBSEARCHDB_SOLR_BASIC_AUTH_USERNAME: ""
      NBSEARCHDB_SOLR_BASIC_AUTH_PASSWORD: ""
      NBSEARCHDB_S3_ENDPOINT_URL: http://nbsearch-minio:9000
      NBSEARCHDB_S3_ACCESS_KEY: nbsearchak
      NBSEARCHDB_S3_SECRET_KEY: nbsearchsk
      NBSEARCHDB_S3_REGION_NAME: ""
      NBSEARCHDB_S3_BUCKET_NAME: notebooks
      NBSEARCHDB_SOLR_NOTEBOOK: jupyter-notebook
      NBSEARCHDB_SOLR_CELL: jupyter-cell
      NBSEARCHDB_MY_SERVER_URL: http://localhost:${LC4RI_PORT}/
      NBSEARCHDB_BASE_DIR: /home/${LC4RI_USER}/notebooks
      NBSEARCHDB_AUTO_UPDATE: "1"

  nbsearch-solr:
    build:
      context: ./.lc4ri/images/nbsearch-solr
    volumes:
      - nbsearch_solr_data:/var/solr/

  nbsearch-minio:
    image: minio/minio:latest
    environment:
      MINIO_ROOT_USER: nbsearchak
      MINIO_ROOT_PASSWORD: nbsearchsk
    command:
      - server
      - /data
    volumes:
      - nbsearch_minio_data:/data
`
}
