package robottxt

import (
	"net/url"
	"testing"
)

func TestParseRobotsTxt(t *testing.T) {

	robotsTxtInstance := &RobotsTxtImpl{
		URL:        "https://test.org",
		disallowed: make(map[string]struct{}),
	}

	robotsTxt := `User-agent: *
Disallow: /cgi-bin/
Disallow: /tmp/
Disallow: /junk/`
	robotsTxtInstance.parseRobotsTxt(robotsTxt)

	if len(robotsTxtInstance.disallowed) != 3 {
		t.Error("Failed to parse all rules")
	}

	checkValueExists(robotsTxtInstance, "/cgi-bin/", t)
	checkValueExists(robotsTxtInstance, "/tmp/", t)
	checkValueExists(robotsTxtInstance, "/junk/", t)
}

func checkValueExists(robotsTxtInstance *RobotsTxtImpl, rule string, t *testing.T) {
	_, ok := robotsTxtInstance.disallowed[rule]
	if !ok {
		t.Errorf("Failed to parse rule: %s", rule)
	}
}
func TestParseRobotsTxtDisallowsAll(t *testing.T) {

	robotsTxtInstance := &RobotsTxtImpl{
		URL:        "https://test.org",
		disallowed: make(map[string]struct{}),
	}

	robotsTxt := `User-agent: *
Disallow: /`
	robotsTxtInstance.parseRobotsTxt(robotsTxt)

	if len(robotsTxtInstance.disallowed) != 1 {
		t.Error("Failed to parse all rules")
	}

	checkValueExists(robotsTxtInstance, "/", t)
}
func TestParseRobotsTxtAllowsAll(t *testing.T) {

	robotsTxtInstance := &RobotsTxtImpl{
		URL:        "https://test.org",
		disallowed: make(map[string]struct{}),
	}

	robotsTxt := `User-agent: *
Disallow: `
	robotsTxtInstance.parseRobotsTxt(robotsTxt)

	if len(robotsTxtInstance.disallowed) != 0 {
		t.Error("Failed to parse all rules")
	}

}

func TestIsDisallowed(t *testing.T) {

	disallowedMap := make(map[string]struct{})

	disallowedMap["/tmp"] = struct{}{}
	robotsTxtInstance := &RobotsTxtImpl{
		URL:        "https://test.org",
		disallowed: disallowedMap,
	}

	URL, _ := url.Parse("https://test.org/test/index.html")
	ok := robotsTxtInstance.IsDisallowed(URL)
	if ok {
		t.Errorf("Failed to allow URL: %s", URL.String())
	}

	URL, _ = url.Parse("https://test.org/tmp/index.html")
	ok = robotsTxtInstance.IsDisallowed(URL)
	if !ok {
		t.Errorf("Failed to disallow URL: %s", URL.String())
	}
	URL, _ = url.Parse("https://test.org/tmp/")
	ok = robotsTxtInstance.IsDisallowed(URL)
	if !ok {
		t.Errorf("Failed to disallow URL: %s", URL.String())
	}
}
