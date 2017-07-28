package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type clientRequest struct {
	endpoint   string
	httpClient boshhttp.HTTPClient

	logger boshlog.Logger
}

func (r clientRequest) Get(path string, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	resp, err := r.httpClient.Get(url)
	if err != nil {
		return bosherr.WrapErrorf(err, "Performing request GET '%s'", url)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bosherr.WrapError(err, "Reading response")
	}

	if resp.StatusCode != http.StatusOK {
		msg := "Responded with non-successful status code '%d' response '%s'"
		return bosherr.Errorf(msg, resp.StatusCode, respBody)
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling response")
	}

	return nil
}

func (r clientRequest) Post(path string, body []byte, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	logTag := "agent.clientRequest"

	r.logger.Debug(logTag, "Performing agent request to POST '%s'", url)

	httpResponse, err := r.httpClient.Post(url, body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Performing request POST '%s'", url)
	}

	defer httpResponse.Body.Close()

	b, err := httputil.DumpResponse(httpResponse, true)
	if err == nil {
		r.logger.Debug(logTag, "Dumping client response:\n%s", string(b))
	}

	responseBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Reading response")
	}

	if httpResponse.StatusCode != http.StatusOK {
		return bosherr.Errorf(
			"Responded with non-successful status code '%d' response '%s'",
			httpResponse.StatusCode, responseBody)
	}

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling response")
	}

	return nil
}
