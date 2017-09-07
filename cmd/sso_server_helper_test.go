package cmd_test

import (
	"fmt"
	"net/http"
	"time"
)

type (
	sso_server struct {
		Values map[string]interface{}
	}
)

func (s *sso_server) login_form(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "cookiename", Value: someCookie, Expires: expiration}
	http.SetCookie(w, &cookie)
	fmt.Fprintln(w, `<html>any html</hmtl>`)
}

func (s *sso_server) valid_saml(w http.ResponseWriter, r *http.Request) {
	s.collect_request_data(r)
	fmt.Fprintln(w, `<input name="SAMLResponse" value="`+validSamlValue+`"/>`)
}

func (s *sso_server) invalid_password_response(w http.ResponseWriter, r *http.Request) {
	s.collect_request_data(r)
	fmt.Fprintln(w, `<div><form><label id="errorText" for="">
		Incorrect user ID or password. Type the correct user ID and password, and try again.
		</label></form></div>`)
}

func (s *sso_server) collect_request_data(r *http.Request) {
	s.Values["parseErr"] = r.ParseForm()
	s.Values["user"] = first(r.PostForm["UserName"])
	s.Values["password"] = first(r.PostForm["Password"])
	var cookie, err = r.Cookie("cookiename")
	if err == nil {
		s.Values["posted-cookie"] = cookie.Value
	}
}

func (s *sso_server) handler(get, post http.HandlerFunc) http.HandlerFunc {
	s.Values = make(map[string]interface{})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if get != nil {
				get(w, r)
			}
		case "POST":
			if post != nil {
				post(w, r)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
