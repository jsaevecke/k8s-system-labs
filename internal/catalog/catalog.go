package catalog

import (
	"fmt"
	"os"
	"path/filepath"

	yamlloader "github.com/jsaevecke/k8s-system-labs/internal/loader/yaml"
)

const definitionFile = "lab.yaml"

type Catalog struct {
	Root string
}

func Find(getwd func() (string, error)) (*Catalog, error) {
	startDirectory, err := getwd()
	if err != nil {
		return nil, fmt.Errorf("read working directory: %w", err)
	}

	current, err := filepath.Abs(startDirectory)
	if err != nil {
		return nil, fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		labsRoot := filepath.Join(current, "labs")
		if info, err := os.Stat(labsRoot); err == nil && info.IsDir() {
			return &Catalog{Root: labsRoot}, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return nil, fmt.Errorf("find labs directory from %q", startDirectory)
		}
		current = parent
	}
}

func (c *Catalog) Resolve(name string) (string, error) {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
		return "", fmt.Errorf("invalid lab name %q", name)
	}

	path := filepath.Join(c.Root, name, definitionFile)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("find lab %q: %w", name, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("lab definition is not a regular file: %s", path)
	}
	return path, nil
}

func (c *Catalog) List() (Labs, error) {
	entries, err := os.ReadDir(c.Root)
	if err != nil {
		return nil, fmt.Errorf("read labs directory: %w", err)
	}

	labs := make(Labs, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(c.Root, entry.Name(), definitionFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("read lab %q: %w", entry.Name(), err)
		}

		definition, err := yamlloader.Load(path)
		if err != nil {
			return nil, fmt.Errorf("load lab %q: %w", entry.Name(), err)
		}
		labs = append(labs, definition)
	}
	return labs, nil
}
