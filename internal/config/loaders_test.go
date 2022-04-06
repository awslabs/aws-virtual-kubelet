/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package config

import (
	"testing"
)

// TestDirectLoader_load tests the DirectLoader's raw configuration loading
func TestDirectLoader_load(t *testing.T) {
	type fields struct {
		DirectConfig ProviderConfig
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// test cases
		{
			name: "Empty Config",
			fields: fields{
				DirectConfig: ProviderConfig{},
			},
			wantErr: false,
		},
		{
			name: "Partial Config",
			fields: fields{
				DirectConfig: ProviderConfig{
					HealthConfig: HealthConfig{
						UnhealthyThresholdCount: 10,
					},
					VKVMAgentConnectionConfig: VkvmaConfig{
						MinConnectTimeoutSeconds: 20,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dl := &DirectLoader{
				DirectConfig: tt.fields.DirectConfig,
			}
			if err := dl.load(); (err != nil) != tt.wantErr {
				t.Errorf("load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// NOTE all loaders share the same validation logic

// TestDirectLoader_validate tests the DirectLoader's validation
func TestDirectLoader_validate(t *testing.T) {
	type fields struct {
		DirectConfig ProviderConfig
	}
	type args struct {
		pc *ProviderConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Empty config",
			fields: fields{},
			args: args{
				pc: &ProviderConfig{},
			},
			wantErr: true,
		},
		{
			name:   "Nonempty config",
			fields: fields{},
			args: args{
				pc: &ProviderConfig{
					ManagementSubnet: ".",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dl := &DirectLoader{
				DirectConfig: tt.fields.DirectConfig,
			}
			if err := dl.validate(tt.args.pc); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileLoader_load(t *testing.T) {
	type fields struct {
		ConfigFilePath string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid config file",
			fields: fields{
				// TODO(guicejg): use os.path or whatever to make this portable
				ConfigFilePath: "../../test/data/config/validConfig.json",
			},
			wantErr: false,
		},
		{
			name: "Missing config file",
			fields: fields{
				ConfigFilePath: "missingFile.json",
			},
			wantErr: true,
		},
		{
			name: "Invalid JSON",
			fields: fields{
				ConfigFilePath: "../../test/data/config/invalidFile.json",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := &FileLoader{
				ConfigFilePath: tt.fields.ConfigFilePath,
			}
			if err := fl.load(); (err != nil) != tt.wantErr {
				t.Errorf("load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileLoader_validate(t *testing.T) {
	type fields struct {
		ConfigFilePath string
	}
	type args struct {
		pc *ProviderConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid config",
			args: args{
				pc: &ProviderConfig{
					ManagementSubnet: ".",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := &FileLoader{
				ConfigFilePath: tt.fields.ConfigFilePath,
			}
			if err := fl.validate(tt.args.pc); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// NOTE both validate functions above delegate logic to this function so most validation tests go here
func Test_validate(t *testing.T) {
	type args struct {
		pc *ProviderConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid config",
			args: args{
				pc: &ProviderConfig{
					ManagementSubnet: ".",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing required field(s)",
			args: args{
				pc: &ProviderConfig{
					HealthConfig: HealthConfig{
						UnhealthyThresholdCount: 10,
					},
					VKVMAgentConnectionConfig: VkvmaConfig{
						MinConnectTimeoutSeconds: 20,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid config with Warm Pool",
			args: args{

				pc: &ProviderConfig{
					ManagementSubnet: ".",
					WarmPoolConfig: []WarmPoolConfig{
						{
							ImageID:      "ami-badf005ba117ab1e5",
							InstanceType: "m72.ginormous",
							Subnets: []string{
								"sg-badf005ba117ab1e5",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Warm Pool config with missing required values",
			args: args{

				pc: &ProviderConfig{
					ManagementSubnet: ".",
					WarmPoolConfig:   []WarmPoolConfig{{}},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validate(tt.args.pc); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
