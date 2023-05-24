/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package utils

import (
	"strings"
)

// TrimmedStringSplit returns a list of strings with their whitespaces removed from the head or tail of the string.
func TrimmedStringSplit(input string, sep string) []string {
	interimString := strings.Split(input, sep)
	for i, substring := range interimString {
		interimString[i] = strings.Trim(substring, " ")
	}
	return interimString
}
