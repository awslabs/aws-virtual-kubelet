/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os/exec"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/cert"
	"k8s.io/klog"
)

const (
	lifeCyclePort = ":8200"
	bootstrapPort = ":8300"
)

type UserData struct {
	VmInit         string `json:"vm-init-config"`
	BootstrapAgent string `json:"bootstrap-agent-config"`
	PreSignedURL   string `json:"bootstrap-agent-download-url"`
	CACertificate  string `json:"bootstrap-agent-ca-cert"`
}

// createFile saves pod spec to tmp folder on VM
func createFile(pod corev1.Pod) error {
	podSpecJson, _ := json.Marshal(pod.Spec)
	err := ioutil.WriteFile("/tmp/pod.json", podSpecJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

// decodeUserData decodes provided data
func decodeUserData(userData string) []byte {
	decode, err := b64.StdEncoding.DecodeString(userData)
	if err != nil {
		klog.Error(err)
	}
	return decode
}

// func getInstanceId get the instanaceid from ec2, this function is not currently used.
func getInstanceId() (string, error) {
	//find instanceId of the ec2
	instanceId, err := exec.Command("bash", "-c", "ec2-metadata --instance-id | cut -d \" \" -f 2").Output()
	if err != nil {
		klog.Error(err)
		return "", err
	}
	klog.Info("instanceId - :", string(instanceId))
	return string(instanceId), err
}

// func getUserData gets ec2 userdata
func getUserData() (*UserData, error) {
	//find user-data of the ec2
	userDataRaw, err := exec.Command("bash", "-c", "curl http://169.254.169.254/latest/user-data").Output()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	//decode the userData  b64.StdEncoding.EncodeToString
	dataDecode := decodeUserData(string(userDataRaw))
	var userData UserData
	json.Unmarshal(dataDecode, &userData)
	return &userData, nil
}

// createSelfSignedCert creates self signed certificate and key
func createSelfSignedCert() ([]byte, []byte, error) {
	certBytes, keyBytes, err := cert.GenerateSelfSignedCertKey("", nil, nil)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}
	return certBytes, keyBytes, nil
}
