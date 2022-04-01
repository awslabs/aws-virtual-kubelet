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
	"context"
	"fmt"
	"net/http"

	"github.com/virtual-kubelet/virtual-kubelet/log"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

// ServeMuxWithAuth implements api.ServerMux
type ServeMuxWithAuth struct {
	auth AuthInterface
	ctx  context.Context
	mux  *http.ServeMux
}

// NewServeMuxWithAuth initiate an instance for ServeMuxWithAuth
func NewServeMuxWithAuth(ctx context.Context, auth AuthInterface) *ServeMuxWithAuth {
	mux := http.NewServeMux()
	return &ServeMuxWithAuth{
		auth: auth,
		ctx:  ctx,
		mux:  mux,
	}
}

// Handle enables auth filter for mux Handle
func (s *ServeMuxWithAuth) Handle(path string, h http.Handler) {
	if s.auth == nil {
		s.mux.Handle(path, h)
	} else {
		s.mux.Handle(path, s.authHandler(h))
	}
}

func (s *ServeMuxWithAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// authHandler is the handlder to authenticate & authorize the request
func (s ServeMuxWithAuth) authHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		info, ok, err := s.auth.AuthenticateRequest(req)
		if err != nil {
			log.G(s.ctx).Infof("Unauthorized, err: %s, RequestURI:%s, UserAgent:%s", err, req.RequestURI, req.UserAgent())
			resp.WriteHeader(http.StatusUnauthorized)
			resp.Write([]byte("Unauthorized"))

			return
		}
		if !ok {
			log.G(s.ctx).Infof("Unauthorized, ok: %t, RequestURI:%s, UserAgent:%s", ok, req.RequestURI, req.UserAgent())
			resp.WriteHeader(http.StatusUnauthorized)
			resp.Write([]byte("Unauthorized"))

			return
		}

		attrs := s.auth.GetRequestAttributes(info.User, req)
		decision, _, err := s.auth.Authorize(req.Context(), attrs)
		if err != nil {
			msg := fmt.Sprintf("Authorization error (user=%s, verb=%s, resource=%s, subresource=%s, err=%s)", attrs.GetUser().GetName(), attrs.GetVerb(), attrs.GetResource(), attrs.GetSubresource(), err)
			log.G(s.ctx).Info(msg)
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(msg))
			return
		}
		if decision != authorizer.DecisionAllow {
			msg := fmt.Sprintf("Forbidden (user=%s, verb=%s, resource=%s, subresource=%s, decision=%d)", attrs.GetUser().GetName(), attrs.GetVerb(), attrs.GetResource(), attrs.GetSubresource(), decision)
			log.G(s.ctx).Info(msg)
			resp.WriteHeader(http.StatusForbidden)
			resp.Write([]byte(msg))
			return
		}

		h.ServeHTTP(resp, req)
	})
}
