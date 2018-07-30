package robottxt

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type RobotsTxt interface {
	IsDisallowed(*url.URL) bool
}

type RobotsTxtImpl struct {
	URL        string
	disallowed map[string]struct{}
}

func New(URL string) RobotsTxt {
	rImpl := &RobotsTxtImpl{
		URL: URL,
	}
	rImpl.init()
	return rImpl
}

func (rb *RobotsTxtImpl) init() {

	URL, err := url.Parse(rb.URL)
	rb.disallowed = make(map[string]struct{})

	response, err := http.Get(fmt.Sprintf("%s://%s/robots.txt", URL.Scheme, URL.Host))
	if err != nil {
		log.Printf("Unable to fetch robots.txt due to %s", err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("Unable to parse robots.txt due to %s", err)
		}
		rb.parseRobotsTxt(string(contents))
	}
}

func (rb *RobotsTxtImpl) parseRobotsTxt(robots string) {

	lines := strings.Split(robots, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Disallow:") {
			disallowedRule := strings.Replace(line, "Disallow:", "", -1)
			disallowedRule = strings.Trim(disallowedRule, " \t\n")
			if disallowedRule == "" {
				continue
			}
			rb.disallowed[disallowedRule] = struct{}{}
		}
	}

}

func (rb *RobotsTxtImpl) IsDisallowed(URL *url.URL) bool {

	disallowed := false
	for disallowedURL, _ := range rb.disallowed {
		if strings.HasPrefix(URL.Path, disallowedURL) {
			disallowed = true
		}
	}
	return disallowed
}
