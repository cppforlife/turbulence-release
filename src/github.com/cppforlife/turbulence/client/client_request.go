package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type ClientRequest struct {
	endpoint     string
	client       string
	clientSecret string
	httpClient   boshhttp.HTTPClient
	logger       boshlog.Logger
}

func NewClientRequest(
	endpoint string,
	client string,
	clientSecret string,
	httpClient boshhttp.HTTPClient,
	logger boshlog.Logger,
) ClientRequest {
	return ClientRequest{
		endpoint:     endpoint,
		client:       client,
		clientSecret: clientSecret,
		httpClient:   httpClient,
		logger:       logger,
	}
}

func (r ClientRequest) Get(path string, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Accept", "application/json")
		req.SetBasicAuth(r.client, r.clientSecret)
	}

	resp, err := r.httpClient.GetCustomized(url, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Performing request GET '%s'", url)
	}

	respBody, err := r.readResponse(resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling response")
	}

	return nil
}

func (r ClientRequest) Post(path string, request interface{}, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(r.client, r.clientSecret)
	}

	reqBytes, err := json.Marshal(request)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling request")
	}

	resp, err := r.httpClient.PostCustomized(url, reqBytes, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Performing request POST '%s'", url)
	}

	respBody, err := r.readResponse(resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling response")
	}

	return nil
}

func (r ClientRequest) readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading response")
	}

	if resp.StatusCode != http.StatusOK {
		msg := "Turbulence responded with non-successful status code '%d' response '%s'"
		return nil, bosherr.Errorf(msg, resp.StatusCode, respBody)
	}

	return respBody, nil
}
