package google

import (
	"backend/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
)

var ACCESS_TOKEN_URL = "https://oauth2.googleapis.com/token"

func GetAccessToken(authorizationCode string, config utils.GoogleConfig) (*AccessTokenResponse, error) {
	slog.Info("getting access token for auth code ", "authorizationCode", authorizationCode)

	var accessTokenResponseData AccessTokenResponse

	err := utils.HttpPostFormData(ACCESS_TOKEN_URL, map[string]string{}, map[string]string{
		"client_id":     config.CLIENT_ID,
		"client_secret": config.CLIENT_SECRET,
		"grant_type":    "authorization_code",
		"code":          authorizationCode,
		"redirect_uri":  config.REDIRECT_URI,
	}, &accessTokenResponseData)
	slog.Info("access Token Response ", "accessTokenResponseData", accessTokenResponseData)

	if err != nil || accessTokenResponseData.AccessToken == "" {
		slog.Error("could not get google access token")
		return nil, err
	}

	return &accessTokenResponseData, nil
}

func UserInformation(accessToken string, config utils.GoogleConfig) (*UserInfoResponse, error) {
	// Create the request URL
	url := "https://www.googleapis.com/oauth2/v2/userinfo"

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}

	// Add the Authorization header with the access token
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info, status: %d, response: %s", resp.StatusCode, string(body))
	}

	// Parse the response body into the UserInfo struct
	var userInfo UserInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to parse user info: %v", err)
	}

	return &userInfo, nil
}
