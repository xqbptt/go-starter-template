package oci

import (
	"backend/storage"
	"backend/utils"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"
)

type OCIStorage struct {
	host          string
	namespace     string
	compartmentID string
	privateKey    *rsa.PrivateKey
	keyID         string
	bucketName    string
	baseURL       string
	parPrefix     string
}

func NewOciStorage(ociConfig utils.OciStorageConfig) (storage.Storage, error) {
	baseURL := fmt.Sprintf("https://%s", ociConfig.HOST)
	privateKey, err := parsePrivateKeyFromPEM(ociConfig.PRIVATE_KEY)
	if err != nil {
		return nil, err
	}

	return &OCIStorage{
		ociConfig.HOST, ociConfig.NAMESPACE, ociConfig.COMPARTMENT_ID, privateKey, ociConfig.KEY_ID, ociConfig.BUCKET_NAME, baseURL, ociConfig.PAR_PREFIX,
	}, nil
}

func (o *OCIStorage) UploadObject(data []byte, objectPath string) (string, error) {
	endpointPath := "/n/" + o.namespace + "/b/" + o.bucketName + "/o/" + objectPath + "?compartmentId=" + o.compartmentID
	_, err := o.executeRequest(http.MethodPut, endpointPath, data)
	if err != nil {
		return "", err
	}
	return objectPath, nil
}

func (o *OCIStorage) GetFullUrl(path string) string {
	return fmt.Sprintf("%s%s", o.parPrefix, path)
}

func (o *OCIStorage) ListAll() (results []string, err error) {
	endpointPath := "/n/" + o.namespace + "/b/" + o.bucketName + "/o" + "?compartmentId=" + o.compartmentID

	body, err := o.executeRequest(http.MethodGet, endpointPath, nil)
	if err != nil {
		return
	}
	var response ListItemsResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}

	for _, item := range response.Objects {
		results = append(results, item.Name)
	}
	return
}

func (o *OCIStorage) DeleteObject(path string) (err error) {
	endpointPath := "/n/" + o.namespace + "/b/" + o.bucketName + "/o/" + path + "?compartmentId=" + o.compartmentID
	log.Println(endpointPath)
	_, err = o.executeRequest(http.MethodDelete, endpointPath, nil)
	return
}

func (o *OCIStorage) GeneratePresignedURL(objectPath string, expiration time.Duration) (string, error) {
	// Create PAR endpoint path
	endpointPath := fmt.Sprintf("/n/%s/b/%s/p/?compartmentId=%s",
		o.namespace,
		o.bucketName,
		o.compartmentID,
	)

	// Prepare PAR request body
	parRequest := struct {
		Name        string `json:"name"`
		AccessType  string `json:"accessType"`
		ObjectName  string `json:"objectName"`
		TimeExpires string `json:"timeExpires"`
	}{
		Name:        "direct-upload",
		AccessType:  "AnyObjectWrite",
		ObjectName:  objectPath,
		TimeExpires: time.Now().Add(expiration).UTC().Format(time.RFC3339Nano),
	}

	requestBody, err := json.Marshal(parRequest)
	if err != nil {
		slog.Error("failed to marshal PAR request", "error", err)
		return "", fmt.Errorf("failed to marshal PAR request: %w", err)
	}

	// Execute PAR creation request
	body, err := o.executeRequest(http.MethodPost, endpointPath, requestBody)
	if err != nil {
		slog.Error("PAR creation failed", "error", err)
		return "", fmt.Errorf("PAR creation failed: %w", err)
	}

	// Parse response
	var parResponse struct {
		AccessUri string `json:"accessUri"`
	}
	if err := json.Unmarshal(body, &parResponse); err != nil {
		slog.Error("failed to parse PAR response", "error", err)
		return "", fmt.Errorf("failed to parse PAR response: %w", err)
	}

	return fmt.Sprintf("%s%s%s", o.baseURL, parResponse.AccessUri, objectPath), nil
}
