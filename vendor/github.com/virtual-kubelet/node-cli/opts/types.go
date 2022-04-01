// Copyright © 2020 The virtual-kubelet authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opts

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Authorization holds the state related to the authorization in the kublet.
type Authorization struct {
	// webhook contains settings related to Webhook authorization.
	Webhook WebhookAuthorization
}

// WebhookAuthorization holds the state related to the Webhook
// Authorization in the Kubelet.
type WebhookAuthorization struct {
	// cacheAuthorizedTTL is the duration to cache 'authorized' responses from the webhook authorizer.
	CacheAuthorizedTTL metav1.Duration
	// cacheUnauthorizedTTL is the duration to cache 'unauthorized' responses from the webhook authorizer.
	CacheUnauthorizedTTL metav1.Duration
}

// Authentication holds the Kubetlet Authentication setttings.
type Authentication struct {
	// webhook contains settings related to webhook bearer token authentication
	Webhook WebhookAuthentication
}

// WebhookAuthentication contains settings related to webhook authentication
type WebhookAuthentication struct {
	// enabled allows bearer token authentication backed by the tokenreviews.authentication.k8s.io API
	Enabled bool
	// cacheTTL enables caching of authentication results
	CacheTTL metav1.Duration
}
