package cmd_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NearlyUnique/awsSts/cmd"
)

const (
	validUser       = "any-user"
	validPassword   = "any-password"
	invalidPassword = "invalid-password"
	validSamlValue  = "any-valid-saml"
	someCookie      = "some-cookie-value"
)

func Test_user_can_log_in(t *testing.T) {
	server := sso_server{}
	ts := httptest.NewServer(server.handler(
		server.login_form,
		server.valid_saml))
	defer ts.Close()

	sso := cmd.SSO{
		Client: &http.Client{},
		URL:    ts.URL,
	}

	saml, err := sso.SingleSignOn(validUser, validPassword)

	equals(t, nil, server.Values["parseErr"])
	ok(t, err)
	equals(t, validUser, server.Values["user"])
	equals(t, validPassword, server.Values["password"])
	equals(t, someCookie, server.Values["posted-cookie"])
	equals(t, validSamlValue, saml.AsAssertion())
}

func Test_logion_errors_are_detected(t *testing.T) {
	server := sso_server{}
	ts := httptest.NewServer(server.handler(
		server.login_form,
		server.invalid_password_response))
	defer ts.Close()

	sso := cmd.SSO{
		Client: &http.Client{},
		URL:    ts.URL,
	}

	_, err := sso.SingleSignOn(validUser, invalidPassword)

	equals(t, "Incorrect user ID or password. Type the correct user ID and password, and try again.", err.Error())
}
