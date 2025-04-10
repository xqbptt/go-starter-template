package utils

import (
	"backend/dto"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

func HttpGet(baseURL string, queryParams map[string]string, responseData any) (err error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return
	}
	q := u.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = dto.NewErrorWithStatus(resp.StatusCode, "request could not be completed successfully")
		return
	}

	// slog.Info("httpGet", "response", string(body))
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		err = dto.NewError("invalid response got from downstream api call")
		return
	}
	return
}

func HttpGetWithHeaders(baseURL string, queryParams map[string]string, headers map[string]string, responseData any) (err error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return
	}

	q := u.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode >= 400 {
		err = dto.NewErrorWithStatus(resp.StatusCode, "request could not be completed successfully")
		return
	}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		err = dto.NewError("invalid response got from downstream api call")
		return
	}

	return
}

func HttpPost(urlStr string, queryParams map[string]string, body any, responseData any) (err error) {
	return HttpPostWithAuth(urlStr, queryParams, body, responseData, "")
}

func HttpPostWithAuth(urlStr string, queryParams map[string]string, body any, responseData any, authToken string) (err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	query := u.Query()
	for key, value := range queryParams {
		query.Add(key, value)
	}
	u.RawQuery = query.Encode()

	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return
		}
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = dto.NewErrorWithStatus(resp.StatusCode, "reqest could not be completed successfully")
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	slog.Info("httpPost", "response", string(respBody))
	json.Unmarshal(respBody, &responseData)

	return
}

func HttpPostFormData(urlStr string, queryParams map[string]string, form map[string]string, responseData any) (err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	query := u.Query()
	for key, value := range queryParams {
		query.Add(key, value)
	}
	u.RawQuery = query.Encode()

	formData := url.Values{}

	for key, value := range form {
		formData.Set(key, value)
	}

	formDataString := formData.Encode()
	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(formDataString))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	slog.Info("httpPostFormData response ", "response", string(respBody))

	if resp.StatusCode >= 400 {
		err = dto.NewErrorWithStatus(resp.StatusCode, "reqest could not be completed successfully")
		return
	}
	json.Unmarshal(respBody, &responseData)

	return
}
