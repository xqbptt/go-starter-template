package oci

type ListItemsResponse struct {
	Objects []struct {
		Name string `json:"name"`
	} `json:"objects"`
}
