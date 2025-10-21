package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			auth := r.Header.Get("Authorization")

			if auth == "" {
				app.InvalidBasicAuthorization(w, r, fmt.Errorf("no auth found"))
				return
			}

			auth_list := strings.SplitN(auth, " ", 2)

			if len(auth_list) != 2 || auth_list[0] != "Basic" {
				app.InvalidBasicAuthorization(w, r, fmt.Errorf("malformed Auth"))
				return
			}

			// decode it
			decoded, err := base64.StdEncoding.DecodeString(auth_list[1])
			if err != nil {
				app.InvalidBasicAuthorization(w, r, err)
				return
			}

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(auth_list) != 2 ||
				creds[0] != app.config.authConfig.basicAuth.username ||
				creds[1] != app.config.authConfig.basicAuth.password {
				app.InvalidBasicAuthorization(w, r, fmt.Errorf("incorrect username/pass"))
				return
			}

			next.ServeHTTP(w, r)

		})
	}

}
