package main

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func Test_error_code_can_be_found(t *testing.T) {
	html := `
<form>
  <div id="error" class="fieldMargin error smallText">
    <label id="errorText" for="">
    Incorrect user ID or password.
    </label>
  </div>
</form>`

	r := strings.NewReader(html)
	doc, _ := goquery.NewDocumentFromReader(r)

	errorText := loginErrorText(doc)

	equals(t, "Incorrect user ID or password.", errorText)
}

func Test_expired_login_code_can_be_found(t *testing.T) {
	html := `
<form>
	<div class="groupMargin" style="display:&#39;&#39;">
			<span id="expiredNotification">You must update your password because your password has expired.</span>
	</div>
</form>`

	r := strings.NewReader(html)
	doc, _ := goquery.NewDocumentFromReader(r)

	errorText := loginErrorText(doc)

	equals(t, "You must update your password because your password has expired.", errorText)
}
