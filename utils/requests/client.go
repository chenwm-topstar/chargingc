package requests

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/go-querystring/query"

	"cchome-admin-topstar/utils/lg"

	"github.com/ernesto-jimenez/httplogger"
	"github.com/pkg/errors"
)

var (
	Local             *http.Client
	Proxy             *http.Client
	healthyNetworkEnv bool
	healthyOnce       sync.Once

	proxyHTTPTransport http.RoundTripper
)

// const proxyAddr = "http://cm:EPdDAU@proxy.momoso.com:3128"

func init() {
	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	// proxyURL, _ := url.Parse(proxyAddr)
	// httpTransport.Proxy = http.ProxyURL(proxyURL)
	httpTransport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	proxyHTTPTransport = httplogger.NewLoggedTransport(httpTransport, &httpLogger{})

	Proxy = &http.Client{Transport: proxyHTTPTransport}
	Proxy.Timeout = 10 * time.Second
	Proxy.CheckRedirect = modifiedCheckRedirect

	Local = &http.Client{}
	Local.Timeout = 10 * time.Second
	Local.CheckRedirect = modifiedCheckRedirect

}

func GetProxyHttpTransport() http.RoundTripper {
	if healthyNetworkEnv {
		return http.DefaultTransport
	}
	return proxyHTTPTransport
}

func GetSmartClient() *http.Client {
	return GetClient(!healthyNetworkEnv)
}

func GetClient(useProxy bool) *http.Client {
	if useProxy {
		return Proxy
	} else {
		return Local
	}
}

func Get(ctx context.Context, url string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	return req.WithContext(ctx)
}

func GetStruct(ctx context.Context, url string, response interface{}) (err error) {
	resp, err := GetSmartClient().Do(Get(ctx, url))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respByt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(respByt, response); err != nil {
		return
	}
	return nil
}

type httpLogger struct{}

func (l *httpLogger) LogRequest(req *http.Request) {
	lg.Debug("[Request]", req.Method, req.URL.String())
}

func (l *httpLogger) LogResponse(req *http.Request, res *http.Response, err error, duration time.Duration) {
	if err != nil {
		lg.Debug("[Response]", err, req.URL.String())
		return
	}
	lg.Debug("[Response]", req.Method, res.StatusCode, duration.Seconds(), req.URL.String())
}

func modifiedCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 15 {
		return errors.New("stopped after 15 redirects")
	}
	return nil
}

func ReadBytes(resp *http.Response, err error) ([]byte, error) {
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Read response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("[%d]%s:%s", resp.StatusCode, resp.Status, string(data))
	}

	return data, nil
}

func ReadJson(resp *http.Response, err error, i interface{}) error {
	data, err := ReadBytes(resp, err)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, i); err != nil {
		lg.Debugf("Decode error:\n%s", string(data))
		return errors.Wrap(err, "Decode response")
	}

	return nil
}

func ExistsUrl(remoteURL string) bool {
	resp, err := http.Head(remoteURL)
	if err != nil {
		lg.Debug("ExistsUrl failed", err.Error(), " image ", remoteURL)
		return false
	}
	if resp.StatusCode != 200 {
		lg.Debug("ExistsUrl failed resp.StatusCode ", resp.StatusCode, " image ", remoteURL)
	}
	return resp.StatusCode == 200
}

func Post(posturl string, req interface{}, headers ...url.Values) (resp *http.Response, err error) {
	reqByt, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrapf(err, "json marshal")
	}

	request, err := http.NewRequest("POST", posturl, bytes.NewReader(reqByt))
	if err != nil {
		return nil, errors.Wrapf(err, "new request")
	}
	if len(headers) > 0 {
		for k, vs := range headers[0] {
			request.Header[k] = vs
		}
	}
	if v := request.Header.Get("Content-Type"); v == "" {
		request.Header.Set("Content-Type", "json")
	}

	return GetSmartClient().Do(request)
}

func PostStruct(posturl string, req interface{}, response interface{}, headers ...url.Values) (err error) {
	resp, err := Post(posturl, req, headers...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.Errorf("httppost status:[%d][%s]", resp.StatusCode, resp.Status)
	}
	if response != nil {
		respByt, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrapf(err, "read all")
		}

		if err = json.Unmarshal(respByt, response); err != nil {
			return errors.Wrapf(err, "json unmarshal")
		}
	}

	return
}

func PostForm(posturl string, req interface{}, headers ...url.Values) (resp *http.Response, err error) {
	vals, err := query.Values(req)
	if err != nil {
		return nil, errors.Wrapf(err, "query values")
	}

	request, err := http.NewRequest("POST", posturl, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, errors.Wrapf(err, "newrequest")
	}
	if len(headers) > 0 {
		for k, vs := range headers[0] {
			request.Header[k] = vs
		}
	}
	if v := request.Header.Get("Content-Type"); v == "" {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return GetSmartClient().Do(request)
}

func PostFormStruct(posturl string, req interface{}, response interface{}, headers ...url.Values) error {
	resp, err := PostForm(posturl, req, headers...)
	if err != nil {
		return errors.Wrapf(err, "postfrom")
	}
	defer resp.Body.Close()
	if response == nil {
		return nil
	}
	return xml.NewDecoder(resp.Body).Decode(response)
}

func Delete(posturl string, req interface{}, response interface{}, headers ...url.Values) (err error) {

	return
}
