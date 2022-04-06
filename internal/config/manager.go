/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package config

import (
	"errors"

	"github.com/creasty/defaults"
	"k8s.io/klog/v2"
)

//type Manager interface {
//	loadConfig(l *Loader)
//	validateConfig()
//}

// InitConfig initializes the global config object given a config loader
func InitConfig(loader Loader) error {
	if loader == nil {
		return errors.New("loader cannot be nil")
	}

	err := loader.load()
	if err != nil {
		klog.ErrorS(err, "Config load failed")
		return err
	}

	if err = defaults.Set(config); err != nil {
		klog.Errorf("Error settings default config values: %v", err)
		return err
	}

	return validate(config)
}
