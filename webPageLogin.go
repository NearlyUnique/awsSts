package main

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func webPageLogin(options *options) *http.Response {
	client, resp, err := getLoginPageCookies(options)
	exitErr(err, "Unable to request login page")

	form, err := loginDetails(options.username, options.password)
	exitErr(err, "Unable to get login details")

	resp, err = postForm(client, options.targetURL, form)
	exitErr(err, "Unable to POST (%s) details", options.targetURL)
	return resp
}
func getLoginPageCookies(options *options) (*http.Client, *http.Response, error) {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	resp, err := client.Get(options.targetURL)
	if options.dumpWork && err != nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			dumpFile("login-form-", body)
		}
		// don't quit because of diagnostics
		err = nil
	}
	return client, resp, err
}
func postForm(client *http.Client, targetURL string, form map[string]string) (*http.Response, error) {
	f := url.Values{}
	for k, v := range form {
		f.Add(k, v)
	}
	req, err := http.NewRequest("POST", targetURL, strings.NewReader(f.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "Prepare login form")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	return resp, errors.Wrap(err, "Retrieving login form")
}
