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
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/howeyc/gopass"
	"github.com/pkg/errors"
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
	verbose     = false
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
	vb := flag.Bool("verbose", true, "Verbose output")
	usage := flag.Usage
	flag.Usage = func() {
		fmt.Printf("AWS STS temporary credentials helper\n"+
			"The following environment variables will be read\n"+
			"\t%s - Url for STS login\n"+
			"\t%s - Username to loing with\n"+
			"\t%s - Password to login with\n", urlEnv, userEnv, passEnv)
		usage()
	}
	flag.Parse()
	if *v {
		fmt.Println(appVersion)
		os.Exit(0)
	}
	profileName = *p
	dumpXML = *dx
	verbose = *vb

	targetURL, err := getLoginURL()
	exitErr(err, "%s environment variable missing", urlEnv)

	client, resp, err := getLoginPageCookies(targetURL)
	exitErr(err, "Unable to request login page")

	form, err := loginDetails()
	exitErr(err, "Unable to get loging details")

	resp, err = postForm(client, targetURL, form)
	exitErr(err, "Unable to POST (%s) details", targetURL)

	xml, assertion, err := getSaml(resp)
	exitErr(err, "Unable to read SAML")

	arns, err := extractArns(xml)
	exitErr(err, "Failed to parse arns")

	arn, err := selectRole(arns)
	exitErr(err, "Failed to select role")

	err = getTempConfigValues(*arn, assertion)
	exitErr(err, "Failed to update config")
}
func getLoginURL() (string, error) {
	targetURL := os.Getenv(urlEnv)
	if len(targetURL) == 0 {
		return "", errors.New("Missing URL")
	}
	log("Checking env..\n%s=%s\n", urlEnv, targetURL)
	return targetURL, nil
}
func getLoginPageCookies(targetURL string) (*http.Client, *http.Response, error) {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	resp, err := client.Get(targetURL)
	return client, resp, err
}
func postForm(client *http.Client, targetURL string, form map[string]string) (*http.Response, error) {
	f := url.Values{}
	for k, v := range form {
		f.Add(k, v)
	}
	req, err := http.NewRequest("POST", targetURL, strings.NewReader(f.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "Prepare loging form")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	return resp, errors.Wrap(err, "Retrieving loging form")
}
func getSaml(resp *http.Response) (xml []byte, assertion string, err error) {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, "", errors.Wrap(err, "Prepare login form for query INPUT elements")
	}
	inputs := []string{}
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		if err != nil {
			return
		}
		name := attrOrEmpty(s, "name")
		if name == "SAMLResponse" {
			assertion = attrOrEmpty(s, "value")
			xml, err = base64.StdEncoding.DecodeString(assertion)
		}
		inputs = append(inputs, name)
	})
	if len(xml) == 0 {
		rawHTML, _ := doc.Html()
		fmt.Fprintf(os.Stderr, "------------------------\n%v------------------------\n", rawHTML)
	}
	return xml, assertion, errors.Wrap(err, "Base64 decoding 'assertion' XML")
}
func loginDetails() (form map[string]string, err error) {
	user := os.Getenv(userEnv)
	pass := os.Getenv(passEnv)
	if len(user) == 0 {
		user, err = getUsername()
	} else {
		log("Username (from %s)='%s'\n", userEnv, user)
	}
	if err == nil && len(pass) == 0 {
		pass, err = getPassword()
	} else {
		log("Password (from %s), length='%d'\n", passEnv, len(pass))
	}
	return map[string]string{
		keyUsername: user,
		keyPassword: pass,
	}, err //error already wrapped
}
func getUsername() (string, error) {
	fmt.Print("User:")
	name, err := bufio.NewReader(os.Stdin).ReadString('\n')
	return name, errors.Wrap(err, "Reading username")
}
func getPassword() (string, error) {
	fmt.Print("Password:")
	pass, err := gopass.GetPasswd()
	return string(pass), errors.Wrap(err, "Reading password")
}
func attrOrEmpty(s *goquery.Selection, name string) string {
	if r, ok := s.Attr(name); ok {
		return r
	}
	return ""
}
func getTempConfigValues(arn Arn, assertion string) error {
	sess, err := session.NewSession()
	if err != nil {
		return errors.Wrap(err, "AWS session")
	}
	svc := sts.New(sess)
	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  aws.String(arn.principal),
		RoleArn:       aws.String(arn.role),
		SAMLAssertion: aws.String(assertion),
	}
	resp, err := svc.AssumeRoleWithSAML(params)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to assume role:%s", arn.role))
	}
	return updateAwsConfig(
		*resp.Credentials.AccessKeyId,
		*resp.Credentials.SecretAccessKey,
		*resp.Credentials.SessionToken)
}
func updateAwsConfig(key, secret, session string) error {
	home := "~"
	for _, env := range []string{"HOME", "USERPROFILE", "HOMEPATH"} {
		home = os.Getenv(env)
		if len(home) > 0 {
			break
		}
	}
	iniFile := path.Join(home, ".aws", "credentials")

	cfg, err := ini.Load(iniFile)
	if err != nil {
		return errors.Wrapf(err, "Loading credentials file %s", iniFile)
	}
	s, err := cfg.GetSection(profileName)
	if err != nil {
		s, err = cfg.NewSection(profileName)
	}
	if err != nil {
		return errors.Wrapf(err, "Unable to add or create section '%s' in  %s", profileName, iniFile)
	}
	s.Key("aws_access_key_id").SetValue(key)
	s.Key("aws_secret_access_key").SetValue(secret)
	s.Key("aws_session_token").SetValue(session)
	err = cfg.SaveTo(iniFile)
	if err != nil {
		return errors.Wrapf(err, "Saving credentials file %s", iniFile)
	}

	log("Profile %q updated\n", profileName)
	return nil
}
func extractArns(raw []byte) ([]Arn, error) {
	response := Response{}
	err := xml.Unmarshal(raw, &response)
	if dumpXML {
		if err != nil {
			fmt.Printf("-------XML--------\n%s\n------------------\n", string(raw))
		} else {
			xml.MarshalIndent(response, "", "  ")
		}
	}
	if err != nil {
		return nil, errors.Wrap(err, "Reading SAML XML")
	}
	roles := []Arn{}
	for _, a := range response.Assertion {
		if a.isRole() {
			roles = a.arns()
			break
		}
	}
	if len(roles) == 0 {
		return nil, errors.New("Expected Role 'https://aws.amazon.com/SAML/Attributes/Role'")
	}

	return roles, nil
}
func selectRole(roles []Arn) (*Arn, error) {
	if len(roles) == 1 {
		log("Using Role %q\n", roles[0].role)
		return &roles[0], nil
	}
	fmt.Println("Select role;")
	for i, a := range roles {
		fmt.Printf("%d: %s\n", i, a.role)
	}
	fmt.Print("?")
	num := -1
	_, err := fmt.Scanf("%d", &num)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read user role selection")
	}
	if num < 0 || num >= len(roles) {
		return nil, errors.Errorf("The nuber %d was not offered", num)
	}
	fmt.Println(num)
	return &roles[num], nil
}
func exitErr(err error, msg string, args ...interface{}) {
	if err != nil {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		fmt.Fprintf(os.Stderr, msg+"\nErr:%v\n", err)
		os.Exit(1)
	}
}
func (a AttributeValue) isRole() bool {
	return a.Name == "https://aws.amazon.com/SAML/Attributes/Role"
}
func (a AttributeValue) arns() []Arn {
	const providerKey = "saml-provider"
	result := []Arn{}
	for _, value := range a.Value {
		parts := strings.Split(value, ",")
		if len(parts) == 2 {
			if strings.Index(parts[0], providerKey) >= 0 {
				result = append(result, Arn{parts[0], parts[1]})
			}
			if strings.Index(parts[1], providerKey) >= 0 {
				result = append(result, Arn{parts[1], parts[0]})
			}
		}
	}
	return result
}
func log(message string, args ...interface{}) {
	if verbose {
		fmt.Printf(message, args...)
	}
}
