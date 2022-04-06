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
	"testing"
)

// BrokenLoader implements a non-functioning loader for testing
type BrokenLoader struct {
}

// load the config given a BrokenLoader (fail to load actually...)
func (bl *BrokenLoader) load() error { return errors.New("this loader is intentionally broken") }
func (bl *BrokenLoader) validate(pc *ProviderConfig) error {
	return errors.New("this loader doesn't validate")
}

func TestInitConfig(t *testing.T) {
	type args struct {
		loader Loader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// NOTE could not find a way to generate an error from the defaults loader (need to modify the struct tags)
		{
			name: "valid Loader",
			args: args{
				loader: &DirectLoader{
					DirectConfig: ProviderConfig{
						ManagementSubnet: ".",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nil Loader",
			args: args{
				loader: nil,
			},
			wantErr: true,
		},
		{
			name: "'empty' Loader",
			args: args{
				loader: *(new(Loader)),
			},
			wantErr: true,
		},
		{
			name: "Broken Loader",
			args: args{
				loader: &BrokenLoader{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitConfig(tt.args.loader); (err != nil) != tt.wantErr {
				t.Errorf("InitConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
