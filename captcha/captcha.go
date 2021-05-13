package captcha

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

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
	ErrorLogger func(message string, err error)
}

func (c *Invisible) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		switch r.Method {
		case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
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

				var form url.Values
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
		ip = splits[0]
	}

	return ip
}
