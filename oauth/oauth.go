package oauth

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"
)

func assignData(data []byte) {
	t := struct {
		ClientID string `yaml:"ClientID"`
		Secret   string `yaml:"Secret"`
	}{}

	yaml.Unmarshal(data, &t)
	ClientID = t.ClientID
	Secret = t.Secret

	return
}

func getSecretsFromEnv() bool {
	var ok bool

	if ClientID, ok = os.LookupEnv("SlackClientID"); !ok {
		return false
	}

	if Secret, ok = os.LookupEnv("SlackSecret"); !ok {
		return false
	}

	return true
}

func configure() {
	var (
		OptLoc string = "/opt/slacker/.secrets.yml"
	)

	CurrentUser, err := user.Current()

	if err != nil {
		fmt.Println("Could not find user")
	}

	UserLoc := fmt.Sprintf("%s%s", CurrentUser.HomeDir, "/.slacksecrets.yml")

	fmt.Println("Looking in environmentals for secrets")
	if ok := getSecretsFromEnv(); ok {
		fmt.Println("Located secrets in environmentals")
		return
	} else {
		fmt.Println("Environmentals not set!")
	}

	if data, err := ioutil.ReadFile(OptLoc); err != nil {
		fmt.Printf("Could not locate %s. \n", OptLoc)
	} else {
		assignData(data)
		return
	}

	if data, err := ioutil.ReadFile(UserLoc); err != nil {
		fmt.Println("Error opening file...")
	} else {
		assignData(data)
		return
	}
}

func NewSlack() *Slack {
	configure()

	var (
		err          error
		AuthorizeURL string = "https://slack.com/oauth/authorize"
		AccessURL    string = "https://slack.com/api/oauth.access"
	)

	s := new(Slack)

	s.Authorize, err = url.Parse(AuthorizeURL)
	if err != nil {
		fmt.Println("Failed creating authorize url.")
	}

	s.Access, err = url.Parse(AccessURL)
	if err != nil {
		fmt.Println("Failed creating access url.")
	}

	return s
}

type Slack struct {
	Authorize *url.URL
	Access    *url.URL
	state     string
}

func generateState() string {
	return fmt.Sprintf("%s%s", Secret, "Test")
}

// Create the url to that the user needs to be redirected to sign into slack.
func (s *Slack) RedirectURL(scopes []string) string {
	s.state = generateState()

	u := s.Authorize
	q := u.Query()
	q.Set("client_id", ClientID)
	q.Set("state", s.state)

	for _, scope := range scopes {
		q.Add("scope", scope)
	}

	u.RawQuery = q.Encode()

	return u.String()
}

func (s *Slack) RequestHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.RedirectURL([]string{"client"}), 303)
}

func (s *Slack) RequestToken(code string, state string) (*http.Response, error) {
	var err error
	if state != s.state {
		err = errors.New("Could not make request")
		return new(http.Response), err
	}

	u := s.Access
	q := u.Query()
	q.Add("code", code)
	u.RawQuery = q.Encode()

	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		err = errors.New("Could not make request")
		return new(http.Response), err
	}

	request.SetBasicAuth(ClientID, Secret)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Println("Failed to make the request")
	}

	return response, err
}

func (s *Slack) ResponseHandler(w http.ResponseWriter, resp *http.Request) {
	q := resp.URL.Query()

	r, err := s.RequestToken(q.Get("code"), q.Get("state"))

	if err != nil {
		fmt.Println("Could not get token")
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Problem reading data", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

var (
	ClientID string
	salt     string = "salt"
	Secret   string
)
