package collector

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
)

func Get(url string, params FlexSwitchParams) ([]byte, error) {
	req := &http.Request{}
	client := &http.Client{}
	err := errors.New("")
	if params.Proto == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
		req, err = http.NewRequest("GET", url, nil)
		if (params.Username != "") && (params.Password != "") {
			req.SetBasicAuth(params.Username, params.Password)
		}
	} else {
		client = &http.Client{}
		req, err = http.NewRequest("GET", url, nil)
		if (params.Username != "") && (params.Password != "") {
			req.SetBasicAuth(params.Username, params.Password)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	htmlBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return htmlBody, err
}
