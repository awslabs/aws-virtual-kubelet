/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package config

// Ec2Provider configuration defaults.
const (
	DefaultOperatingSystem = "Linux"
	DefaultCpuCapacity     = "20"
	DefaultMemoryCapacity  = "40Gi"
	DefaultStorageCapacity = "40Gi"
	DefaultPodCapacity     = "200"
	DefaultConfigLocation  = "/etc/config/config.json"
)

// ExtendedConfig contains additional configuration collected from CLI flags that is not part of VK's InitConfig
type ExtendedConfig struct {
	KubeConfigPath string
}

// ProviderConfig represents the entire configuration structure
type ProviderConfig struct {
	// AWS Region to launch and interact with resources in
	Region string `default:"us-west-2"`
	// Name of the cluster used to tag ENIs (nodes)
	ClusterName string `default:"aws-virtual-kubelet"`
	// Subnet used for... TODO describe how this is used
	ManagementSubnet string

	VMConfig                  VMConfig       `default:"{}"`
	BootstrapAgent            BootstrapAgent `default:"{}"`
	HealthConfig              HealthConfig
	VKVMAgentConnectionConfig VkvmaConfig
	WarmPoolConfig            []WarmPoolConfig `default:"-"`
	//http client request timeout limit
	AWSClientTimeoutSeconds int `default:"10"`
	//http client dialer timeout limit
	AWSClientDialerTimeoutSeconds int `default:"2"`
}

// VMConfig defines Default configurations for EC2 Instances if not otherwise specified.
type VMConfig struct {
	// Default AMI to launch pods on (can be overridden in pod spec)
	DefaultAMI string
	// Base64 EC2 user data for... TODO describe how this is used
	InitData string
}

// BootstrapAgent defines where to locate the Bootstrap Agent information during VM startup
type BootstrapAgent struct {
	// S3 Bucket name to download VKVMAgent binary from
	S3Bucket string
	// S3 Bucket key pointing to VKVMAgent binary
	S3Key string
	// GRPC (VKVMAgent) port to connect to
	GRPCPort int `default:"8200"`
	// Base64 EC2 user data for... TODO describe how this is used
	InitData string
}

// WarmPoolConfig represents the contents of WarmPool feature configuration, which is optional.
type WarmPoolConfig struct {
	// Desired number of warm pool instances to maintain
	DesiredCount int `default:"10"`
	// Instance profile to associate with warm pool instances
	IamInstanceProfile string
	// Security groups to set on warm pool instances
	SecurityGroups []string
	// Key pair used to launch warm pool instances
	KeyPair string
	// AMI ID to use for creation of warm pool instances
	ImageID string
	// Instance type to use for warm pool instances
	InstanceType string
	// Subnets to launch warm pool instances in
	Subnets []string
}

// HealthConfig contains podMonitor health monitoring settings and defaults
type HealthConfig struct {
	// Consecutive failure results required before reporting unhealthy status back to provider
	UnhealthyThresholdCount int `default:"5"`
	// [UNUSED] maximum failure results before forcing provider to take action
	UnhealthyMaxCount int `default:"30"`
	// Frequency to conduct "active" (polling) health checks at
	HealthCheckIntervalSeconds int `default:"60"`
}

// VkvmaConfig contains VKVMAgent connection and related settings
type VkvmaConfig struct {
	Port int `default:"8200"`
	// NOTE ⚠️  Initial connection timeout is independent from reconnect params below (set values accordingly)
	TimeoutSeconds int `default:"300"`
	// Minimum time to give a connection to complete
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

// package-level config "singleton" for global static access (provided via Config function below)
var config *ProviderConfig

// Config provides an exported accessor allowing callers to retrieve configuration
func Config() *ProviderConfig {
	return config
}
