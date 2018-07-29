package fetcher

import (
	"fmt"
	"testing"
)

func TestIsExternalLink(t *testing.T) {

	fetchr := &BasicFetcher{baseDomain: "www.monzo.com"}

	err := testLink("https://itunes.apple.com/gb/app/mondo/id1052238659", true, fetchr)
	if err != nil {
		t.Error(err)
	}

	err = testLink("https://web.monzo.com", false, fetchr)
	if err != nil {
		t.Error(err)
	}

	err = testLink("/login", false, fetchr)
	if err != nil {
		t.Error(err)
	}

	err = testLink("#inpage-ref", true, fetchr)
	if err != nil {
		t.Error(err)
	}

	err = testLink("tel:+442038720620", true, fetchr)
	if err != nil {
		t.Error(err)
	}

}

func TestNormalise(t *testing.T) {
	fetchr := &BasicFetcher{baseDomain: "monzo.com"}

	normalisedURL, err := fetchr.normalise("/login", "https://monzo.com/login")

	if normalisedURL.String() != "https://monzo.com/login" || err != nil {
		t.Error()
	}

	normalisedURL, err = fetchr.normalise("https://www.monzo.com/login", "https://monzo.com")

	if normalisedURL.String() != "https://www.monzo.com/login" || err != nil {
		t.Error()
	}
}

func TestRun(t *testing.T) {
	inputC := make(chan string)
	outputC := make(chan *URLMap)
	fetchr := &BasicFetcher{baseDomain: "monzo.com", inputChannel: inputC, outputChannel: outputC}
	go fetchr.Run()

	go func() {
		inputC <- "https://monzo.com"
	}()

	urlMap := <-outputC
	if urlMap.URL != "https://monzo.com" {
		t.Fail()
	}

}

func testLink(url string, expectedOutcome bool, fetchr *BasicFetcher) error {

	realOutcome := fetchr.isExternalOrInvalidLink(url, "https://www.monzo.com")
	if realOutcome != expectedOutcome {
		return fmt.Errorf("Test for %s failed", url)
	}
	return nil
}
