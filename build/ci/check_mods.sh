#!/bin/sh

# This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
# © 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
#
# This AWS Content is provided subject to the terms of the AWS Customer Agreement
# available at http://aws.amazon.com/agreement or other written agreement between
# Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.

set -e

exit_code=0

make mod
git diff --exit-code go.mod go.sum || exit_code=$?

if [ ${exit_code} -eq 0 ]; then
	exit 0
fi

echo "please run \`make mod\` and check in the changes"
exit ${exit_code}
