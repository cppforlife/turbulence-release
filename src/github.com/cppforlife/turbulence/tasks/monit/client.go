package monit

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	gourl "net/url"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/http"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type httpClient struct {
	host, username, password string
	client                   boshhttp.Client
	logger                   boshlog.Logger
}

func NewHTTPClient(
	host, username, password string,
	client boshhttp.Client,
	logger boshlog.Logger,
) Client {
	return httpClient{
		host:     host,
		username: username,
		password: password,
		client:   client,
		logger:   logger,
	}
}

func (c httpClient) Services() ([]Service, error) {
	var services []Service

	status, err := c.status()
	if err != nil {
		return services, bosherr.WrapError(err, "Getting status from monit")
	}

	for _, service := range status.Services.Services {
		// skip system service which does not have a PID (not a process)
		if service.PID != 0 {
			services = append(services, Service{Name: service.Name, PID: service.PID})
		}
	}

	return services, nil
}

func (c httpClient) status() (status, error) {
	statusURL := gourl.URL{
		Scheme:   "http",
		Host:     c.host,
		Path:     "/_status2",
		RawQuery: "format=xml",
	}

	resp, err := c.makeGETRequest(statusURL)
	if err != nil {
		return status{}, bosherr.WrapError(err, "Sending status request to monit")
	}

	defer resp.Body.Close()

	respBody, err := c.validateResponse(resp)
	if err != nil {
		return status{}, bosherr.WrapError(err, "Getting monit status")
	}

	// todo cheat a bit instead of loading charset library
	respBodyStr := strings.Replace(string(respBody), ` encoding="ISO-8859-1"`, "", 1)

	var st status

	err = xml.Unmarshal([]byte(respBodyStr), &st)
	if err != nil {
		return status{}, bosherr.WrapError(err, "Unmarshalling monit status")
	}

	return st, nil
}

func (c httpClient) validateResponse(resp *http.Response) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, bosherr.Errorf("Request failed with status '%s'", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading body of failed monit response")
	}

	return body, nil
}

func (c httpClient) makeGETRequest(target gourl.URL) (*http.Response, error) {
	request, err := http.NewRequest("GET", target.String(), nil)
	if err != nil {
		return nil, err
	}

	request.SetBasicAuth(c.username, c.password)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.client.Do(request)
}
