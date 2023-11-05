package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func GetIndex(ctx context.Context) (url string, data []byte, err error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			url = req.URL.String()
			return nil
		},
		Transport: NewContextTransport(ctx),
	}

	r, err := client.Get("https://www.southparkstudios.com/")
	if err != nil {
		return url, data, fmt.Errorf("could not get southparkstudios.com: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return url, data, fmt.Errorf("could not get southparkstudios.com: %s", r.Status)
	}

	data, err = io.ReadAll(r.Body)
	if err != nil {
		return url, data, fmt.Errorf("could not read response body: %w", err)
	}

	return url, data, nil
}

func InitialUrl(indexUrl string, data []byte) (u string, err error) {
	iu, err := url.ParseRequestURI(indexUrl)
	if err != nil {
		return "", err
	}

	/*
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
		if err != nil {
			return "", fmt.Errorf("could not parse from index %s html: %w", indexUrl, err)
		}
	*/

	/*
		lang, exists := doc.Find("html").First().Attr("lang")
		if !exists {
			return "", fmt.Errorf("could not find html lang attribute")
		}

	*/
	iu.Path = "/episodes/940f8z/south-park-cartman-gets-an-anal-probe-season-1-ep-1"

	return iu.String(), nil
}

func StartingUrl(ctx context.Context) (string, error) {
	url, data, err := GetIndex(ctx)
	if err != nil {
		return "", err
	}

	url, err = InitialUrl(url, data)
	if err != nil {
		return "", err

	}

	return url, nil
}
