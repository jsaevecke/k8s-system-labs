package yaml

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
	yamlv3 "gopkg.in/yaml.v3"
)

func Load(path string) (domain.Lab, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return domain.Lab{}, fmt.Errorf("resolve lab path: %w", err)
	}

	file, err := os.Open(absolutePath)
	if err != nil {
		return domain.Lab{}, fmt.Errorf("open lab definition: %w", err)
	}
	defer file.Close()

	var definition domain.Lab
	decoder := yamlv3.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&definition); err != nil {
		return domain.Lab{}, fmt.Errorf("decode lab definition: %w", err)
	}
	definition.SourcePath = absolutePath

	if err := definition.Validate(); err != nil {
		return domain.Lab{}, err
	}
	return definition, nil
}
