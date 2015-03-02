package snappy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"launchpad.net/snappy/helpers"
)

var (
	ubuntuoneAPIBase  = "https://login.ubuntu.com/api/v2"
	ubuntuoneOauthAPI = ubuntuoneAPIBase + "/tokens/oauth"
)

type StoreToken struct {
	OpenID      string `json:"openid"`
	TokenName   string `json:"token_name"`
	DateUpdated string `json:"date_updated"`
	DateCreated string `json:"date_created"`
	Href        string `json:"href"`

	TokenKey       string `json:"token_key"`
	TokenSecret    string `json:"token_secret"`
	ConsumerSecret string `json:"consumer_secret"`
	ConsumerKey    string `json:"consumer_key"`
}

// RequestStoreToken requests a token for accessing the ubuntu store
func RequestStoreToken(username, password, tokenName, otp string) (*StoreToken, error) {
	data := map[string]string{
		"email":      username,
		"password":   password,
		"token_name": tokenName,
	}
	if otp != "" {
		data["otp"] = otp
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", ubuntuoneOauthAPI, strings.NewReader(string(jsonData)))
	req.Header.Set("content-type", "application/json")
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 403:
		return nil, errors.New("invalid credentials")
	case resp.StatusCode != 200 && resp.StatusCode != 201:
		return nil, fmt.Errorf("failed to get store token: %v (%v)", resp.StatusCode, resp)
	}

	var token StoreToken
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// WriteStoreToken takes the token and stores it on the filesystem for
// later reading via ReadStoreToken()
func WriteStoreToken(token StoreToken) error {
	homeDir, err := helpers.CurrentHomeDir()
	if err != nil {
		return err
	}
	targetFile := filepath.Join(homeDir, ".config", "snappy", "sso.json")
	if err := helpers.EnsureDir(filepath.Dir(targetFile), 0700); err != nil {
		return err
	}
	outStr, err := json.MarshalIndent(token, "", " ")
	if err != nil {
		return nil
	}

	return helpers.AtomicWriteFile(targetFile, []byte(outStr), 0600)
}
