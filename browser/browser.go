package browser

import (
	"bytes"
	"net/http"
)

type Browser struct {
	Cookies      []*http.Cookie
	LastResponse http.Response
}

func NewBrowser() *Browser {
	return &Browser{}
}

func (b *Browser) Request(method string, url string, data string) {
	var dataBuffer *Buffer

	if data != "" {
		dataBuffer = bytes.NewBufferString(data)

	}

	r, err := http.NewRequest(method, url, dataBuffer)
	for i, c := range Cookies {
		r.AddCookie(c)
	}

	if err != nil {
		return r, err
	}

	resp, err := http.DefaultTransport.RoundTrip(r)

	if err != nil {
		return r, err
	}

	for i, c = range resp.Cookies {
		append(b.Cookies, c)
	}

	return r, err
}
