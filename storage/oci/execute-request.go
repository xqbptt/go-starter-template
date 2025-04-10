package oci

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func (o *OCIStorage) executeRequest(method string, endpointPath string, data []byte) (body []byte, err error) {
	url := o.baseURL + endpointPath

	var buffer *bytes.Buffer
	req, err := http.NewRequest(method, url, nil)
	if data != nil {
		buffer = bytes.NewBuffer(data)
		req, err = http.NewRequest(method, url, buffer)
	}
	if err != nil {
		return
	}

	headersToSign := []string{"(request-target)", "date", "host"}
	headers := strings.Join(headersToSign, " ")

	currentTime := time.Now().UTC()
	currentTimeString := currentTime.Format("Mon, 02 Jan 2006 15:04:05 GMT")

	hostHeader := `host: ` + o.host
	dateHeader := `date: ` + currentTimeString
	escapedTarget := endpointPath
	requestTargetHeader := fmt.Sprintf("(request-target): %s %s", strings.ToLower(req.Method), escapedTarget)
	signatureString := strings.Join([]string{requestTargetHeader, dateHeader, hostHeader}, "\n")

	signature, err := signStringWithRSA256(signatureString, o.privateKey)
	if err != nil {
		slog.Error("error generating signature:", slog.Any("error", err))
		return
	}

	signatureHeader := fmt.Sprintf(`version="1",keyId="%s",algorithm="rsa-sha256",headers="%s",signature="%s"`, o.keyID, headers, signature)

	req.Header.Add("Date", currentTimeString)
	req.Header.Add("Authorization", "Signature "+signatureHeader)
	if req.Method != http.MethodGet && data != nil {
		if req.Method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "application/octet-stream")
		}
		// req.Header.Add("Content-Type", "application/json")
		// req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-Length", fmt.Sprint(len(data)))
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		err = fmt.Errorf("request failed with status code: %d", resp.StatusCode)
		slog.Error("request failed", slog.Any("error", err))
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return
	}
	return
}
