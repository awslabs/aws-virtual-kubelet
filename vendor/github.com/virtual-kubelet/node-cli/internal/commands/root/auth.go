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

package root

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	clientset "k8s.io/client-go/kubernetes"
	authenticationclient "k8s.io/client-go/kubernetes/typed/authentication/v1"
	authorizationclient "k8s.io/client-go/kubernetes/typed/authorization/v1"

	"github.com/virtual-kubelet/node-cli/opts"
)

const (
	metricsPath = "/metrics"
	statsPath   = "/stats/"
	logsPath    = "/logs/"
)

// AuthInterface contains all methods required by the auth filters
type AuthInterface interface {
	authenticator.Request
	authorizer.RequestAttributesGetter
	authorizer.Authorizer
}

// VirtualKubeletAuth implements AuthInterface
type VirtualKubeletAuth struct {
	// authenticator identifies the user for requests to the Kubelet API
	authenticator.Request
	// authorizerAttributeGetter builds authorization.Attributes for a request to the Kubelet API
	authorizer.RequestAttributesGetter
	// authorizer determines whether a given authorization.Attributes is allowed
	authorizer.Authorizer
}

// NewVirtualKubeletAuth returns a AuthInterface composed of the given authenticator, attribute getter, and authorizer
func NewVirtualKubeletAuth(authenticator authenticator.Request, authorizerAttributeGetter authorizer.RequestAttributesGetter, authorizer authorizer.Authorizer) AuthInterface {
	return &VirtualKubeletAuth{authenticator, authorizerAttributeGetter, authorizer}
}

// BuildAuth creates an authenticator, an authorizer, and a matching authorizer attributes getter compatible with the virtual-kubelet's needs
func BuildAuth(nodeName types.NodeName, client clientset.Interface, config opts.Opts) (AuthInterface, func(<-chan struct{}), error) {
	// Get clients, if provided
	var (
		tokenClient authenticationclient.TokenReviewInterface
		sarClient   authorizationclient.SubjectAccessReviewInterface
	)
	if client != nil && !reflect.ValueOf(client).IsNil() {
		tokenClient = client.AuthenticationV1().TokenReviews()
		sarClient = client.AuthorizationV1().SubjectAccessReviews()
	}

	authenticator, runAuthenticatorCAReload, err := BuildAuthn(tokenClient, config.Authentication, config.ClientCACert)
	if err != nil {
		return nil, nil, err
	}

	attributes := NewNodeAuthorizerAttributesGetter(nodeName)

	authorizer, err := BuildAuthz(sarClient, config.Authorization)
	if err != nil {
		return nil, nil, err
	}

	return NewVirtualKubeletAuth(authenticator, attributes, authorizer), runAuthenticatorCAReload, nil
}

// BuildAuthn creates an authenticator compatible with the virtual-kubelet's needs
func BuildAuthn(client authenticationclient.TokenReviewInterface, authn opts.Authentication, clientCACert string) (authenticator.Request, func(<-chan struct{}), error) {
	var dynamicCAContentFromFile *dynamiccertificates.DynamicFileCAContent
	var err error
	if len(clientCACert) == 0 {
		return nil, nil, errors.New("no ca file is provided, cannot use webhook authorization")
	}
	dynamicCAContentFromFile, err = dynamiccertificates.NewDynamicCAContentFromFile("client-ca-bundle", clientCACert)
	if err != nil {
		return nil, nil, err
	}

	authenticatorConfig := authenticatorfactory.DelegatingAuthenticatorConfig{
		Anonymous:                          false,
		CacheTTL:                           authn.Webhook.CacheTTL.Duration,
		ClientCertificateCAContentProvider: dynamicCAContentFromFile,
	}

	if authn.Webhook.Enabled {
		if client == nil {
			return nil, nil, errors.New("no client provided, cannot use webhook authentication")
		}
		authenticatorConfig.TokenAccessReviewClient = client
	}

	authenticator, _, err := authenticatorConfig.New()
	if err != nil {
		return nil, nil, err
	}

	return authenticator, func(stopCh <-chan struct{}) {
		if dynamicCAContentFromFile != nil {
			go dynamicCAContentFromFile.Run(1, stopCh)
		}
	}, err
}

// BuildAuthz creates an authorizer compatible with the virtual-kubelet's needs
func BuildAuthz(client authorizationclient.SubjectAccessReviewInterface, authz opts.Authorization) (authorizer.Authorizer, error) {
	if client == nil {
		return nil, errors.New("no client provided, cannot use webhook authorization")
	}
	authorizerConfig := authorizerfactory.DelegatingAuthorizerConfig{
		SubjectAccessReviewClient: client,
		AllowCacheTTL:             authz.Webhook.CacheAuthorizedTTL.Duration,
		DenyCacheTTL:              authz.Webhook.CacheUnauthorizedTTL.Duration,
	}
	return authorizerConfig.New()
}

type nodeAuthorizerAttributesGetter struct {
	nodeName types.NodeName
}

// NewNodeAuthorizerAttributesGetter creates a new authorizer.RequestAttributesGetter for the node.
func NewNodeAuthorizerAttributesGetter(nodeName types.NodeName) authorizer.RequestAttributesGetter {
	return nodeAuthorizerAttributesGetter{nodeName: nodeName}
}

// GetRequestAttributes populates authorizer attributes for the requests to the virtual-kubelet API.
// Default attributes are: {apiVersion=v1,verb=<http verb from request>,resource=nodes,name=<node name>,subresource=proxy}
func (n nodeAuthorizerAttributesGetter) GetRequestAttributes(u user.Info, r *http.Request) authorizer.Attributes {
	apiVerb := ""
	switch r.Method {
	case "POST":
		apiVerb = "create"
	case "GET":
		apiVerb = "get"
	case "PUT":
		apiVerb = "update"
	case "PATCH":
		apiVerb = "patch"
	case "DELETE":
		apiVerb = "delete"
	}

	requestPath := r.URL.Path

	attrs := authorizer.AttributesRecord{
		User:            u,
		Verb:            apiVerb,
		Namespace:       "",
		APIGroup:        "",
		APIVersion:      "v1",
		Resource:        "nodes",
		Subresource:     "proxy",
		Name:            string(n.nodeName),
		ResourceRequest: true,
		Path:            requestPath,
	}

	switch {
	case isSubpath(requestPath, statsPath):
		attrs.Subresource = "stats"
	case isSubpath(requestPath, metricsPath):
		attrs.Subresource = "metrics"
	case isSubpath(requestPath, logsPath):
		attrs.Subresource = "log"
	}

	return attrs
}

func isSubpath(subpath, path string) bool {
	path = strings.TrimSuffix(path, "/")
	return subpath == path || (strings.HasPrefix(subpath, path) && subpath[len(path)] == '/')
}
