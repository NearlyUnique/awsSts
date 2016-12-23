package main

import (
	"bufio"
	"encoding/base64"
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/howeyc/gopass"
	"gopkg.in/ini.v1"
)

const (
	urlEnv      = "AWSSTS_URL"
	userEnv     = "AWSSTS_USER"
	passEnv     = "AWSSTS_PASS"
	keyUsername = "UserName"
	keyPassword = "Password"
)

var (
	profileName = "saml"
	dumpXML     = false
)

type (
	// AttributeValue .
	AttributeValue struct {
		Name  string   `xml:"Name,attr"`
		Value []string `xml:"AttributeValue"`
	}
	// Response .
	Response struct {
		XMLName   xml.Name         `xml:"Response"`
		Assertion []AttributeValue `xml:"Assertion>AttributeStatement>Attribute"`
	}
	//Arn principal and role
	Arn struct {
		principal, role string
	}
)

func main() {
	p := flag.String("profile", "saml", "The profile to store these temproary credentials (use 'default' to make it the default)")
	// iniFile := flag.String("credentials", "", "override default credentials file to write to")
	v := flag.Bool("version", false, "Display version")
	dx := flag.Bool("dump-xml", false, "dump XML from AWS")
	u := flag.Usage
	flag.Usage = func() {
		fmt.Printf("AWS STS temporary credentials helper\n"+
			"The following environment variables will be read\n"+
			"\t%s - Url for STS login\n"+
			"\t%s - Username to loing with\n"+
			"\t%s - Password to login with\n", urlEnv, userEnv, passEnv)
		u()
	}
	flag.Parse()
	if *v {
		fmt.Println(appVersion)
		os.Exit(0)
	}
	profileName = *p
	dumpXML = *dx

	targetURL := getLoginURL()
	client, resp := getLoginPageCookies(targetURL)
	resp = postForm(client, targetURL, loginDetails())
	getFormValues(resp)
}
func getLoginURL() string {
	targetURL := os.Getenv(urlEnv)
	if len(targetURL) == 0 {
		exitErr(fmt.Errorf("Missing URL"), "%s environment variable missing", urlEnv)
	}
	fmt.Printf("Checking env..\n%s=%s\n", urlEnv, targetURL)
	return targetURL
}
func getLoginPageCookies(targetURL string) (*http.Client, *http.Response) {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	resp, err := client.Get(targetURL)
	exitErr(err, "Unable to request login page")
	return client, resp
}
func postForm(client *http.Client, targetURL string, form map[string]string) *http.Response {
	f := url.Values{}
	for k, v := range form {
		f.Add(k, v)
	}
	req, err := http.NewRequest("POST", targetURL, strings.NewReader(f.Encode()))
	exitErr(err, "Unable create POST (%s)", targetURL)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	exitErr(err, "Unable to POST (%s) details", targetURL)
	return resp
}
func getFormValues(resp *http.Response) {
	ok := false
	doc, err := goquery.NewDocumentFromResponse(resp)
	exitErr(err, "Unable to parse login details")
	inputs := []string{}
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		name := attrOrEmpty(s, "name")
		if name == "SAMLResponse" {
			v := attrOrEmpty(s, "value")
			xml, err := base64.StdEncoding.DecodeString(v)
			exitErr(err, "Not base64")

			arn, err := extractArns(xml)
			exitErr(err, "Failed to parse arns")

			getTempConfigValues(*arn, v)
			ok = true
		}
		inputs = append(inputs, name)
	})
	if !ok {
		rawHTML, _ := doc.Html()
		fmt.Printf("------------------------\n%v------------------------\n", rawHTML)
		exitErr(fmt.Errorf("Failed to find SAMLResponse input element in inputs %v", inputs), "Failed to parse response form")
	}
	return
}
func loginDetails() map[string]string {
	user := os.Getenv(userEnv)
	pass := os.Getenv(passEnv)
	if len(user) == 0 {
		user = getUsername()
	} else {
		fmt.Printf("Username (from %s)='%s'\n", userEnv, user)
	}
	if len(pass) == 0 {
		pass = getPassword()
	} else {
		fmt.Printf("Password (from %s), length='%d'\n", passEnv, len(pass))
	}
	return map[string]string{
		keyUsername: user,
		keyPassword: pass,
	}
}
func getUsername() string {
	fmt.Print("User:")
	u, err := bufio.NewReader(os.Stdin).ReadString('\n')
	exitErr(err, "Unable to read form stdin")
	return u
}
func getPassword() string {
	fmt.Print("Password:")
	pass, err := gopass.GetPasswd()
	exitErr(err, "Unable to read form stdin")
	return string(pass)
}
func attrOrEmpty(s *goquery.Selection, name string) string {
	if r, ok := s.Attr(name); ok {
		return r
	}
	return ""
}
func getTempConfigValues(arn Arn, assertion string) {
	sess, err := session.NewSession()
	exitErr(err, "failed to create session")
	svc := sts.New(sess)
	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  aws.String(arn.principal),
		RoleArn:       aws.String(arn.role),
		SAMLAssertion: aws.String(assertion),
	}
	resp, err := svc.AssumeRoleWithSAML(params)
	exitErr(err, "Unable to assume AWS role %s\n", arn.role)
	updateAwsConfig(
		*resp.Credentials.AccessKeyId,
		*resp.Credentials.SecretAccessKey,
		*resp.Credentials.SessionToken)
}
func updateAwsConfig(key, secret, session string) {
	home := "~"
	for _, env := range []string{"HOME", "USERPROFILE", "HOMEPATH"} {
		home = os.Getenv(env)
		if len(home) > 0 {
			break
		}
	}
	iniFile := path.Join(home, ".aws", "credentials")

	cfg, err := ini.Load(iniFile)
	exitErr(err, "Failed to load shared credentials '%s'", iniFile)
	s, err := cfg.GetSection(profileName)
	if err != nil {
		s, err = cfg.NewSection(profileName)
	}
	exitErr(err, "Failed to find or create new profile (%s) config in '%s'", profileName, iniFile)
	s.Key("aws_access_key_id").SetValue(key)
	s.Key("aws_secret_access_key").SetValue(secret)
	s.Key("aws_session_token").SetValue(session)
	err = cfg.SaveTo(iniFile)
	exitErr(err, "Failed to save shared credentials")

	fmt.Printf("Profile %q updated\n", profileName)
}
func extractArns(raw []byte) (*Arn, error) {
	response := Response{}
	err := xml.Unmarshal(raw, &response)
	if dumpXML {
		if err != nil {
			fmt.Printf("-------XML--------\n%s\n------------------\n", string(raw))
		} else {
			xml.MarshalIndent(response, "", "  ")
		}
	}
	exitErr(err, "failed to parse SAML XML")
	roles := []Arn{}
	for _, a := range response.Assertion {
		if a.isRole() {
			roles = a.arns()
			break
		}
	}
	if len(roles) == 0 {
		return nil, fmt.Errorf("Expected Role 'https://aws.amazon.com/SAML/Attributes/Role'")
	} else if len(roles) == 1 {
		fmt.Printf("Using Role %q\n", roles[0].role)
		return &roles[0], nil
	}
	fmt.Println("Select role;")
	for i, a := range roles {
		fmt.Printf("%d: %s\n", i, a.role)
	}
	num, err := bufio.NewReader(os.Stdin).ReadString('\n')
	exitErr(err, "Unable to read choice")
	i, err := strconv.Atoi(strings.Trim(num, " \r\n"))
	exitErr(err, "That's not a number '%s'", num)
	if i >= len(roles) {
		exitErr(fmt.Errorf("Out of range"), "That's not a valid choice '%s'", num)
	}
	return &roles[i], nil
}
func exitErr(err error, msg string, args ...interface{}) {
	if err != nil {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		fmt.Println(msg+"\nErr:%v\n", err)
		os.Exit(1)
	}
}

func (a AttributeValue) isRole() bool {
	return a.Name == "https://aws.amazon.com/SAML/Attributes/Role"
}
func (a AttributeValue) arns() []Arn {
	const providerKey = "saml-provider"
	result := []Arn{}
	// fmt.Printf("found %d values\n", len(a.Value))
	for _, value := range a.Value {
		// fmt.Printf("\tvalue[%d]=%s\n", i, value)
		parts := strings.Split(value, ",")
		// fmt.Printf("part count %d\n", len(parts))
		if len(parts) == 2 {
			fmt.Printf("\t\tpart[0]=%s ==? %s\n", parts[0], providerKey)
			// fmt.Printf("\t\tpart[1]=%s\n", parts[1])
			if strings.Index(parts[0], providerKey) >= 0 {
				// fmt.Printf("provider:%s\n", parts[0])
				result = append(result, Arn{parts[0], parts[1]})
			}
			if strings.Index(parts[1], providerKey) >= 0 {
				result = append(result, Arn{parts[1], parts[0]})
			}
		}
	}
	return result
}
