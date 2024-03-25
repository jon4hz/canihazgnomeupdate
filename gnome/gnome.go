package gnome

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Extension represents the basic information of a GNOME extension
type Extension struct {
	UUID            string                  `json:"uuid"`
	Name            string                  `json:"name"`
	URL             string                  `json:"url"`
	ShellVersionMap map[string]ShellVersion `json:"shell_version_map"`
}

type ShellVersion struct {
	PK      int64 `json:"pk"`
	Version int64 `json:"version"`
}

// SearchExtension searches GNOME extensions by the given query
func SearchExtension(query string) (*Extension, error) {
	var (
		page     = 1
		result   []Extension
		numPages int
		err      error
	)

	for numPages == 0 || page <= numPages {
		result, numPages, err = search(query, page)
		if err != nil {
			return nil, err
		}

		wanted := extensionByUUID(query, result)
		if wanted != nil {
			return wanted, nil
		}
		page++
	}
	return nil, fmt.Errorf("extension %q not found", query)
}

func extensionByUUID(uuid string, exts []Extension) *Extension {
	for _, ext := range exts {
		if ext.UUID == uuid {
			return &ext
		}
	}
	return nil
}

func search(query string, page int) ([]Extension, int, error) {
	// Prepare the search URL
	searchURL := "https://extensions.gnome.org/extension-query/"
	params := url.Values{}
	params.Add("search", query)
	params.Add("page", fmt.Sprintf("%d", page))
	fullURL := fmt.Sprintf("%s?%s", searchURL, params.Encode())

	// Make the HTTP GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	// The API response structure (simplified)
	var result struct {
		Extensions []Extension `json:"extensions"`
		NumPages   int         `json:"numpages"`
	}

	// Parse the JSON response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}

	return result.Extensions, result.NumPages, nil
}
