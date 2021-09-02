package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/egeneralov/ejabberd_api_exporter/internal/generic/str"
	"net/http"
	"strings"
	"time"
)

type Api struct {
	client   *http.Client
	login    string
	password string
	vhost    string
	endpoint string
}

func New(login, password, vhost, endpoint string) *Api {
	return &Api{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					//RootCAs:            &x509.CertPool{},
					//ClientCAs:          &x509.CertPool{},
					InsecureSkipVerify: true,
				},
				TLSHandshakeTimeout: time.Second * 3,
				DisableKeepAlives:   false,
				DisableCompression:  false,
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     10,
				IdleConnTimeout:     30,
			},
			Timeout: time.Second * 10,
		},
		login:    login,
		password: password,
		vhost:    vhost,
		endpoint: endpoint,
	}
}

func (a *Api) getJson(url string, target interface{}) error {
	r, err := a.client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = r.Body.Close() }()
	return json.NewDecoder(r.Body).Decode(target)
}

func (a *Api) postJson(url string, payload, target interface{}) error {
	j, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	response, err := a.client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	return json.NewDecoder(response.Body).Decode(target)
}

func (a *Api) RegisteredUsers() ([]string, error) {
	var (
		result []string
		pre    []string
		err    error
	)
	err = a.postJson(
		a.endpoint+"/api/registered_users",
		struct {
			Host string `json:"host"`
		}{Host: a.vhost},
		&pre,
	)
	for _, el := range pre {
		if !str.InSlice(el, result) {
			result = append(result, el)
		}
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) ConnectedUsers() ([]string, error) {
	var (
		result []string
		pre    []string
		err    error
	)
	err = a.getJson(
		a.endpoint+"/api/connected_users",
		&pre,
	)
	if err != nil {
		return nil, err
	}
	for _, el := range pre {
		s := strings.Split(el, "@")
		if len(s) > 0 {
			if !str.InSlice(s[0], result) {
				result = append(result, s[0])
			}
		} else {
			result = append(result, el)
		}
	}
	return result, nil
}

func (a *Api) UserResources(username string) ([]string, error) {
	var (
		result []string
		pre    []string
		err    error
	)
	err = a.postJson(
		a.endpoint+"/api/user_resources",
		struct {
			Host string `json:"host"`
			User string `json:"user"`
		}{Host: a.vhost, User: username},
		&pre,
	)
	for _, el := range pre {
		if !str.InSlice(el, result) {
			result = append(result, el)
		}
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Stats get single metric. ["registeredusers", "onlineusers", "onlineusersnode", "uptimeseconds", "processes"]
func (a *Api) Stats(metric string) (int64, error) {
	var (
		result struct {
			Stat int64 `json:"stat"`
		}
		err error
	)
	err = a.postJson(
		a.endpoint+"/api/stats",
		struct {
			Name string `json:"name"`
		}{Name: metric},
		&result,
	)
	if err != nil {
		return -1, err
	}
	return result.Stat, nil
}
