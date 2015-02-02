package browser

import (
	"bytes"
	"fmt"
	"net/http"
)

type Browser struct {
	Cookies      []*http.Cookie
	LastResponse http.Response
	Debug        bool
}

func NewBrowser(debug bool) *Browser {
	return &Browser{Debug: debug}
}

func (b *Browser) Request(method string, url string, data string) (*http.Response, error) {
	dataBuffer := bytes.NewBufferString(data)

	r, err := http.NewRequest(method, url, dataBuffer)
	if err != nil {
		return nil, err
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	for _, c := range b.Cookies {
		r.AddCookie(c)
	}

	if err != nil {
		return nil, err
	}

	if b.Debug {
		b.ShowCookies()
		fmt.Printf("Requesting: %s\n", url)
	}

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	if b.Debug {
		fmt.Printf("Status: %s\nHeaders: %s\n\n", resp.Status, resp.Header)
	}

	if err != nil {
		return resp, err
	}

	for _, c := range resp.Cookies() {
		b.Cookies = append(b.Cookies, c)
	}

	return resp, err
}

func (b *Browser) ShowCookies() {
	fmt.Println("\n\n-------- [COOKIES]")
	for i, c := range b.Cookies {
		fmt.Printf(" Cookie %d - %s\n", i, c.Name)
	}
	fmt.Println("-------- [COOKIES]\n\n")
}
