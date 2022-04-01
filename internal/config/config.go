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
	"io/ioutil"
	"os"
	"reflect"

	"github.com/creasty/defaults"
	"github.com/pkg/errors"
	"github.com/virtual-kubelet/node-cli/provider"
	"k8s.io/klog"
)

// Ec2Provider configuration defaults.
const (
	DefaultOperatingSystem = "Linux"
	DefaultCpuCapacity     = "20"
	DefaultMemoryCapacity  = "40Gi"
	DefaultStorageCapacity = "40Gi"
	DefaultPodCapacity     = "200"
	DefaultConfigLocation  = "/etc/config/config.json"
)

// ProviderConfig represents the contents of the provider configuration file.
type ProviderConfig struct {
	Region           string
	ClusterName      string
	ManagementSubnet string
	NodeName         string
	VMConfig         VMConfig
	BootstrapAgent   BootstrapAgent
	WarmPoolConfig   []WarmPoolConfig
}

// VMConfig defines Default configurations for EC2 Instances if not otherwise specified.
type VMConfig struct {
	DefaultAMI string
	InitData   string
}

// BootstrapAgent defines where to locate the Bootstrap Agent information during VM startup
type BootstrapAgent struct {
	S3Bucket string
	S3Key    string
	GRPCPort int
	InitData string
}

// WarmPoolConfig represents the contents of WarmPool feature configuration, which is optional.
type WarmPoolConfig struct {
	DesiredCount       int      `json:"DesiredCount,omitempty,string"`
	IamInstanceProfile string   `json:"IamInstanceProfile,omitempty"`
	SecurityGroups     []string `json:"Securitygroups,omitempty,string"`
	KeyPair            string   `json:"KeyPair,omitempty"`
	ImageID            string   `json:"ImageID,omitempty"`
	InstanceType       string   `json:"InstanceType,omitempty"`
	Subnets            []string `json:"Subnets,omitempty,string"`
}

// ExtendedConfig contains additional configuration collected from CLI flags that is not part of VK's InitConfig
type ExtendedConfig struct {
	KubeConfigPath string
}

// LoadInitParams loads a ProviderConfig with required values
func LoadInitParams(initCfg provider.InitConfig) (ProviderConfig, error) {
	var err error
	var config ProviderConfig

	fileLocation := initCfg.ConfigPath

	if fileLocation == "" {
		fileLocation = DefaultConfigLocation
	}

	configFile, err := os.Open(fileLocation)
	if err != nil {
		klog.Error("opening Config file: ", err.Error())
		return config, err
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		klog.Error("parsing Config file: ", err.Error())
		return config, err
	}

	// NOTE this parsing approach prevents us from validating specific fields in "subconfigs" (e.g. a field
	//  in WarmPoolConfig can't be marked optional while another in that config is marked required)
	// Check if mandatory fields are missing
	// Replace Reflect implementation of validation with specific field checks if performance is not satisfactory.
	v := reflect.ValueOf(config)
	optionalFields := map[string]bool{
		"NodeName":        true,
		"VmInit":          true,
		"BootstrapAgent":  true,
		"BootstrapBucket": true,
		"BoostrapKey":     true,
		"WarmPoolConfig":  true,
		"HealthConfig":    true,
	}
	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		fieldValue := v.Field(i).Interface()
		if (fieldValue == "") && !optionalFields[fieldName] {
			klog.Errorf("LoadInitParams errored on %s ", fieldName)
			return config, errors.Errorf("mandatory field  %s not defined in Virtual Kubelet Config", fieldValue)
		}
	}
	return config, nil
}

// --------------------------------------------------------------------------------
// NOTE New Config implementation below (is replacing existing ProviderConfig)

// HealthConfig contains podMonitor health monitoring settings and defaults
type HealthConfig struct {
	// consecutive failure results required before reporting unhealthy status back to provider
	UnhealthyThresholdCount int `default:"5"`
	// [UNUSED] maximum failure results before forcing provider to take action
	UnhealthyMaxCount int `default:"30"`
	// frequency to conduct "active" (polling) health checks at
	HealthCheckIntervalSeconds int `default:"60"`
}

// VkvmaConfig contains VKVMAgent connection and related settings
type VkvmaConfig struct {
	// NOTE ⚠️  Initial connection timeout is independent from reconnect params below (set values accordingly)
	TimeoutSeconds int `default:"300"`
	// minimum time to give a connection to complete
	MinConnectTimeoutSeconds   int `default:"60"`
	HealthCheckIntervalSeconds int `default:"60"`
	// See https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md for backoff implementation details
	Backoff struct {
		// BaseDelay is the amount of time to backoff after the first failure.
		BaseDelaySeconds int `default:"1"`
		// Multiplier is the factor with which to multiply backoffs after a
		// failed retry. Should ideally be greater than 1.
		Multiplier float64 `default:"1.5"`
		// Jitter is the factor with which backoffs are randomized.
		Jitter float64 `default:"0.5"`
		// MaxDelay is the upper bound of backoff delay.
		MaxDelaySeconds int `default:"120"`
	}
	Keepalive struct {
		// After a duration of this time if the client doesn't see any activity it
		// pings the server to see if the transport is still alive.
		// If set below 10s, a minimum value of 10s will be used instead.
		TimeSeconds int `default:"60"`
		// After having pinged for keepalive check, the client waits for a duration
		// of Timeout and if no activity is seen even after that the connection is
		// closed.
		TimeoutSeconds int `default:"120"`
	}
}

// providerConfig represents the entire configuration structure
type providerConfig struct {
	// NOTE var names *must* match `config.json` key names
	HealthConfig              HealthConfig
	VKVMAgentConnectionConfig VkvmaConfig
}

// Loader captures input params needed to load configuration
type Loader struct {
	ConfigPath string
}

var config *providerConfig

// LoadConfig is meant to be called _once_ during app init to load the configuration and handle any errors.
// Config() should be use after to obtain any configuration elements
func (cl Loader) LoadConfig() error {
	config = &providerConfig{}

	if err := defaults.Set(config); err != nil {
		klog.Errorf("Error settings default config values: %v", err)
		return err
	}

	if cl.ConfigPath == "" {
		cl.ConfigPath = DefaultConfigLocation
	}

	configFile, err := ioutil.ReadFile(cl.ConfigPath)
	if err != nil {
		klog.Errorf("Error reading Config file: %v", err)
		return err
	}

	// TODO(guicejg): detect missing configuration and notify (or fail if non-optional)
	if err := json.Unmarshal(configFile, config); err != nil {
		klog.Errorf("Error parsing Config file: %v", err)
		return err
	}

	return nil
}

// Config provides an exported accessor allowing callers to retrieve configuration
func Config() *providerConfig {
	return config
}
