package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang/glog"

	"github.com/kubernetes/client-go/kubernetes/typed/authentication/v1"
	authenticationapi "k8s.io/api/authentication/v1"
)

// TokenReviewMiddleware - Middleware to validate a bearer token using k8s
// token review.
type TokenReviewMiddleware struct {
	TokenReview v1.TokenReviewInterface
	Authorizer  UserInfoAuthorizer
}

type osbError struct {
	Description string `json:"description,omitempty"`
}

// Middleware - function that conforms to gorilla-mux middleware.
func (tr TokenReviewMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		glog.V(3).Infof("Request to %v; checking token for authentication", r.RequestURI)
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if auth == "" {
			writeOSBStatusCodeErrorResponse(w, http.StatusUnauthorized, osbError{
				Description: "unable to find authentication token",
			})
			glog.Infof("unable to find the authentication token")
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
			writeOSBStatusCodeErrorResponse(w, http.StatusUnauthorized, osbError{
				Description: "invalid authentication",
			})
			glog.Infof("invalid authentication - %v\n", auth)
			return
		}
		token := parts[1]
		if len(token) == 0 {
			writeOSBStatusCodeErrorResponse(w, http.StatusUnauthorized, osbError{
				Description: "unable to find authentication token",
			})
			glog.Infof("unable to find authentication token- %v\n", token)
			return
		}
		t, err := tr.TokenReview.Create(&authenticationapi.TokenReview{Spec: authenticationapi.TokenReviewSpec{Token: token}})
		if err != nil {
			writeOSBStatusCodeErrorResponse(w, http.StatusUnauthorized, osbError{
				Description: "unable to authenticate token",
			})
			glog.Infof("unable to authenticate token- %v\n", err)
			return
		}
		if !t.Status.Authenticated {
			writeOSBStatusCodeErrorResponse(w, http.StatusUnauthorized, osbError{
				Description: "user was not authenticated",
			})
			glog.Infof("user was not authenticated")
			return
		}
		glog.V(3).Infof("token passed authentication")
		// If UserInfoAuthorizer is defined check that the user is authorized.
		if tr.Authorizer != nil {
			glog.V(3).Infof("Checking user: %v for authorization", t.Status.User)
			decision, err := tr.Authorizer.Authorize(t.Status.User, r)
			if decision == DecisionDeny || decision == DecisionNoOpinion || err != nil {
				writeOSBStatusCodeErrorResponse(w, http.StatusUnauthorized, osbError{
					Description: "unable to authorize user",
				})
				glog.Infof("unable to authorize user: %#+v\n err: %v\n", t.Status.User, err)
				return
			}
			glog.V(3).Infof("user: %v for passed authorization", t.Status.User)
		}
		next.ServeHTTP(w, r)
	})
}

// writeOSBStatusCodeErrorResponse - This is taken from osb-broker-lib.
// In the future I would like to re-use this functionality from there.
func writeOSBStatusCodeErrorResponse(w http.ResponseWriter, statusCode int, osbErr osbError) {
	data, err := json.Marshal(osbErr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	w.Write(data)
}
