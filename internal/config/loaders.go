/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"k8s.io/klog/v2"
)

// Loader is a generic interface describing the functions of a config loader
type Loader interface {
	load() error
	validate(pc *ProviderConfig) error
}

// FileLoader loads the configuration given a path to a file
type FileLoader struct {
	ConfigFilePath string
}

// DirectLoader receives a struct of the expected configuration
type DirectLoader struct {
	DirectConfig ProviderConfig
}

// load the config given a FileLoader
func (fl *FileLoader) load() error {
	configFile, err := ioutil.ReadFile(fl.ConfigFilePath)

	if err != nil {
		klog.ErrorS(err, "Error reading Config file", "file", fl.ConfigFilePath)
		return err
	}

	// init package var config
	config = &ProviderConfig{}

	// populate config with the unmarshalled JSON
	if err = json.Unmarshal(configFile, config); err != nil {
		klog.ErrorS(err, "Error parsing Config file", "file", fl.ConfigFilePath)
		return err
	}

	return nil
}

// load the config given a DirectLoader
func (dl *DirectLoader) load() error {
	// init package var config
	config = &dl.DirectConfig

	return nil
}

// validate with a FileLoader receiver delegates to the static validate function below
func (fl *FileLoader) validate(pc *ProviderConfig) error { return validate(pc) }

// validate with a DirectLoader receiver delegates to the static validate function below
func (dl *DirectLoader) validate(pc *ProviderConfig) error { return validate(pc) }

// validate implements validation logic for all configuration data
func validate(pc *ProviderConfig) error {
	// if config is (an) empty (ProviderConfig struct)
	if reflect.DeepEqual(pc, &ProviderConfig{}) {
		return errors.New("config is empty")
	}

	var errs []string

	if pc.ManagementSubnet == "" {
		errs = append(errs, "ManagementSubnet is required")
	}

	errs = validateWarmPoolConfig(pc, errs)

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %v", strings.Join(errs, ", "))
	}

	return nil
}

// validateWarmPoolConfig checks the warm pool sub-configuration for errors
func validateWarmPoolConfig(pc *ProviderConfig, errs []string) []string {
	// if any warm pool configs are provided, validate required members for each
	if pc.WarmPoolConfig != nil && len(pc.WarmPoolConfig) > 0 {
		for i, wpc := range pc.WarmPoolConfig {
			if wpc.ImageID == "" {
				errs = append(errs, fmt.Sprintf("WarmPoolConfig.ImageID is required for WarmPoolConfig[%d]", i))
			}
			if wpc.InstanceType == "" {
				errs = append(errs, fmt.Sprintf("WarmPoolConfig.InstanceType is required for WarmPoolConfig[%d]", i))
			}
			if len(wpc.Subnets) == 0 {
				errs = append(errs, fmt.Sprintf("WarmPoolConfig.Subnets can't be empty for WarmPoolConfig[%d]", i))
			}
		}
	}
	return errs
}
