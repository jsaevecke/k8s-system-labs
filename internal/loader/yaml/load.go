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

	var definition domain.Lab
	decoder := yamlv3.NewDecoder(file)
	decoder.KnownFields(true)
	decodeErr := decoder.Decode(&definition)
	closeErr := file.Close()
	if decodeErr != nil {
		return domain.Lab{}, fmt.Errorf("decode lab definition: %w", decodeErr)
	}
	if closeErr != nil {
		return domain.Lab{}, fmt.Errorf("close lab definition: %w", closeErr)
	}
	definition.SourcePath = absolutePath

	if err := definition.Validate(); err != nil {
		return domain.Lab{}, err
	}
	return definition, nil
}
