/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package utils

const (
	CertManagerNamespace = "cert-manager"
	VKCaCert             = "vk-ca-cert"
	VKCaCertKey          = "vk-ca-cert-key"
	VKClientCert         = "vk-client-cert"
	VKClientCertKey      = "vk-client-cert-key"
	VKClientCertData     = "vk-client-cert-data"
	VKCaCertData         = "vk-ca-cert-data"
)
