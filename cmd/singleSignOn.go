package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

type (
	//SSO Single Signon
	SSO struct {
		Client *http.Client
		URL    string
	}
)

//SingleSignOn logs the user in to the STS signon page and retreives the content
func (sso SSO) SingleSignOn(username, password string) (*Saml, error) {
	resp, err := sso.getLogonPageCookies()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to request logon page")
	}
	resp, err = sso.postForm(username, password)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to POST (%s) details", sso.URL)
	}
	journal("Signed on as %q via %q", username, sso.URL)
	return extractSaml(resp)
}

func (sso SSO) getLogonPageCookies() (*http.Response, error) {
	jar, _ := cookiejar.New(nil)
	sso.Client.Jar = jar

	resp, err := sso.Client.Get(sso.URL)

	dumpFileFn("logon-form-", func() []byte {
		defer resp.Body.Close()
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte(fmt.Sprintf("`cannot read body, Status:%d", resp.StatusCode))
		}
		err = nil
		return body
	})

	return resp, err
}

func (sso SSO) postForm(username, password string) (*http.Response, error) {
	f := url.Values{
		"UserName": []string{username},
		"Password": []string{password},
	}

	req, err := http.NewRequest("POST", sso.URL, strings.NewReader(f.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "Prepare logon form")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := sso.Client.Do(req)

	return resp, errors.Wrap(err, "Retrieving logon form")
}

func extractSaml(resp *http.Response) (*Saml, error) {
	var err error
	var saml Saml
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "Prepare login form for query INPUT elements")
	}

	dumpFileFn("login-response-", func() []byte {
		rawHTML, _ := doc.Html()
		return []byte(rawHTML)
	})

	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		if err != nil {
			return
		}
		name := attrOrEmpty(s, "name")
		if name == "SAMLResponse" {
			saml = Saml(attrOrEmpty(s, "value"))
		}
	})
	if len(saml) == 0 {
		msg := loginErrorText(doc)
		if len(msg) == 0 {
			msg = "Failed to find SAML xml"
		}
		return nil, errors.New(msg)
	}
	return &saml, errors.Wrap(err, "extracting saml response")
}

func attrOrEmpty(s *goquery.Selection, name string) string {
	if r, ok := s.Attr(name); ok {
		return r
	}
	return ""
}

func loginErrorText(doc *goquery.Document) string {
	for _, id := range []string{"#errorText", "#expiredNotification"} {
		sel := doc.Find(id)
		if sel != nil {
			msg := sel.Text()
			msg = strings.Trim(msg, " \t\r\n")
			if len(msg) > 0 {
				return msg
			}
		}
	}
	return ""
}
