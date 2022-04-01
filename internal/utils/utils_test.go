/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/
package utils

import (
	"testing"
)

type TrimmedStringSplitTest struct {
	original  string
	separator string
	expected  []string
}

func TestTrimmedStringSplit(t *testing.T) {

	cases := []TrimmedStringSplitTest{
		{"subnet-0b36c1d32eb61d6cb", ",", []string{"subnet-0b36c1d32eb61d6cb"}},
		{"subnet-0b36c1d32eb61d6cb, subnet-0b36c1d32eb6fdsa ", ",", []string{"subnet-0b36c1d32eb61d6cb", "subnet-0b36c1d32eb6fdsa"}},
	}

	for _, tt := range cases {
		actual := TrimmedStringSplit(tt.original, tt.separator)
		for i, substring := range actual {
			if substring != tt.expected[i] {
				t.Errorf("TrimmedStringSplit(%s): expected %s, actual %s", tt.original, tt.expected[i], substring)
			}
		}
	}

}
