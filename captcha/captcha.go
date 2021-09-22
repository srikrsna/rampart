package captcha

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/srikrsna/rampart"
)

const (
	ErrInvalidInputResponse = "invalid-input-response"
	ErrMissingInputSecret   = "missing-input-secret"
	ErrInvalidInputSecret   = "invalid-input-secret"
	ErrMissingInputResponse = "missing-input-response"
	ErrBadRequest           = "bad-request"
	ErrTimeoutOrDuplicate   = "timeout-or-duplicate"
)

const recpatchaVerifyUrl = "https://www.google.com/recaptcha/api/siteverify"

type Invisible struct {
	Secret      string
	Rampart     *rampart.Rampart
	KeyFunc     func(r *http.Request) string
	Skip        func(r *http.Request) bool
	ErrorLogger func(message string, err error)
}

func (c *Invisible) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		switch r.Method {
		case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
			if c.Skip(r) {
				break
			}

			key := c.KeyFunc(r)
			canpass, err := c.Rampart.CanPass(ctx, key)
			if err != nil {
				c.ErrorLogger("unable to check can pass", err)
				canpass = true
			}

			if !canpass {
				response := r.Header.Get("X-Captcha-Response")
				if response == "" {
					http.Error(w, "Verify with captcha", http.StatusTooManyRequests)
					return
				}

				ip := IpAsKey(r)

				form := make(url.Values)
				form.Set("secret", c.Secret)
				form.Set("response", response)
				form.Set("remoteip", ip)

				res, err := http.PostForm(recpatchaVerifyUrl, form)
				if err != nil {
					c.ErrorLogger("unable to verify recaptcha token", err)
					http.Error(w, "Unknown Error", http.StatusInternalServerError)
					return
				}
				defer res.Body.Close()

				var cr recaptchaResponse
				if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
					c.ErrorLogger("unable to decode recaptcha response", err)
					http.Error(w, "Unknown Error", http.StatusInternalServerError)
					return
				}

				if !cr.Success {
					for _, ec := range cr.ErrorCodes {
						if ec == ErrInvalidInputResponse || ec == ErrTimeoutOrDuplicate {
							http.Error(w, "Invalid Catpcha", http.StatusBadRequest)
							return
						}
					}

					c.ErrorLogger("error in verification", errors.New(cr.ErrorCodes[0]))
					http.Error(w, "Unknown Error", http.StatusInternalServerError)
					return
				} 

				if err := c.Rampart.Clear(r.Context(), key); err != nil {
					c.ErrorLogger("unable to clear key", err)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

type recaptchaResponse struct {
	Success            bool      `json:"success"`
	ChallengeTimestamp time.Time `json:"challenge_ts"`
	Hostname           string    `json:"hostname"`
	ErrorCodes         []string  `json:"error-codes"`
}

func IpAsKey(r *http.Request) string {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if splits := strings.Split(r.Header.Get("X-Forwarded-For"), ","); splits[0] != "" {
		ip = strings.TrimSpace(splits[0])
	}

	return ip
}

func IpSkip(ips ...string) func(r *http.Request) bool {
	for i := range ips {
		ips[i] = strings.TrimSpace(ips[i])
	}

	return func(r *http.Request) bool {
		ip := IpAsKey(r)
		for _, aip := range ips {
			if ip == aip {
				return true
			}
		}
		return false
	}
}

func PathSkip(patterns ...string) (func(r *http.Request) bool, error) {
	globs := make([]glob.Glob, 0, len(patterns))
	for _, p := range patterns {
		g, err := glob.Compile(p)
		if err != nil {
			return nil, err
		}
		globs = append(globs, g)
	}

	return func(r *http.Request) bool {		
		path := strings.Split(r.URL.Host, ":")[0] + strings.TrimSpace(strings.ToLower(r.URL.Path))
		for _, g := range globs {
			if g.Match(path) {
				return true
			}
		}
		return false
	}, nil
}
