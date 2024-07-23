package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/env/v10"
	"gopkg.in/yaml.v3"
)

const configFileEnvVarName = "CONFIG_FILE"

func ParseEnvsAndFile(cfg interface{}, filePath string) error {
	var resErrors []error

	configFilePath := os.Getenv(configFileEnvVarName)
	if len(configFilePath) > 0 {
		filePath = configFilePath
	}

	err := env.Parse(cfg)

	switch {
	case err != nil && len(filePath) == 0:
		return fmt.Errorf("parsing envs: %w", err)

	case err == nil && len(filePath) == 0:
		return nil

	case err != nil && len(filePath) > 0:
		resErrors = append(resErrors, fmt.Errorf("parsing envs: %w", err))
	}

	bufFile, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	if err = json.Unmarshal(bufFile, cfg); err == nil {
		return nil
	}
	resErrors = append(resErrors, fmt.Errorf("parsing JSON: %w", err))

	if err = yaml.Unmarshal(bufFile, cfg); err == nil {
		return nil
	}
	resErrors = append(resErrors, fmt.Errorf("parsing YAML: %w", err))

	return errors.Join(resErrors...)
}
