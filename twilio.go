package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-playground/form"

	"github.com/pkg/errors"
)

type TwilioClient struct {
	AccountSID, AuthToken string

	Debug bool
}

func NewTwilioClient(sid, token string) *TwilioClient {
	return &TwilioClient{
		AccountSID: sid,
		AuthToken:  token,
	}
}

func (t TwilioClient) Authenticate(req *http.Request) {
	req.SetBasicAuth(t.AccountSID, t.AuthToken)
}

func (t TwilioClient) Request(ctx context.Context, method, url string, params url.Values) ([]byte, error) {
	var contentType string
	var bodyReader io.Reader
	if params != nil {
		contentType = "application/x-www-form-urlencoded"
		bodyReader = strings.NewReader(params.Encode())
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	t.Authenticate(req)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if t.Debug {
		raw, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(raw))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if t.Debug {
		raw, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(raw))
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		return data, errors.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	return data, nil
}

type StartCallRequest struct {
	From       string
	To         string
	SendDigits string
	Url        string
	Method     string
	Record     bool
}

func (t TwilioClient) StartCall(ctx context.Context, req StartCallRequest) ([]byte, error) {
	vals, err := form.NewEncoder().Encode(req)
	if err != nil {
		return nil, err
	}
	return t.Request(ctx, "POST", fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Calls.json", t.AccountSID), vals)
}
