package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

var log = InitLogger()

type alfredCollection struct {
	Items []alfredItem `json:"items"`
}

type Icon struct {
	Path string `json:"path"`
}

type alfredItem struct {
	Arg      string `json:"arg"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Match    string `json:"match"`
	UID      string `json:"uid"`
	Icon     Icon   `json:"icon"`
}

type dashboard struct {
	UID         string `json:"uid"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Type        string `json:"type"`
	IsStarred   bool   `json:"isStarred"`
	FolderTitle string `json:"folderTitle"`
}

func main() {
	grafanaHost := os.Getenv("GRAFANA_HOST")
	query := strings.TrimSpace(os.Args[1])

	apiURL, err := buildAPIURL(grafanaHost)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	req, err := createRequest(apiURL, query)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	resp, err := sendRequest(req)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	defer resp.Body.Close()

	dashboards, err := parseResponse(resp)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	items := buildAlfredItems(dashboards, grafanaHost)
	outputJSON(items)
}

func buildAPIURL(grafanaHost string) (*url.URL, error) {
	apiURL, err := url.Parse(grafanaHost)
	if err != nil {
		return nil, err
	}
	apiURL.Path = path.Join(apiURL.Path, "api/search")
	return apiURL, nil
}

func createRequest(apiURL *url.URL, query string) (*http.Request, error) {
	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, err
	}

	AddAuth(req)
	if query != "" {
		q := req.URL.Query()
		q.Add("query", query)
		req.URL.RawQuery = q.Encode()
	}

	log.Debugf("Requesting: %s", req.URL.String())
	return req, nil
}

func sendRequest(req *http.Request) (*http.Response, error) {
	tr := http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	httpClient := &http.Client{Transport: &tr}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	log.Debugf("Response Status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Response: %d", resp.StatusCode)
	}
	return resp, nil
}

func parseResponse(resp *http.Response) ([]dashboard, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dashboards []dashboard
	err = json.Unmarshal(body, &dashboards)
	if err != nil {
		return nil, err
	}
	return dashboards, nil
}

func buildAlfredItems(dashboards []dashboard, grafanaHost string) []alfredItem {
	var items []alfredItem
	for _, dashboard := range dashboards {
		targetURL, err := url.Parse(grafanaHost)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		targetURL.Path = path.Join(targetURL.Path, dashboard.URL)
		match := strings.ReplaceAll(dashboard.Title, "(", "")
		match = strings.ReplaceAll(match, ")", "")
		match = strings.ReplaceAll(match, "/", "")
		iconFile := "icons/dashboard.svg"
		if dashboard.Type == "dash-folder" {
			iconFile = "icons/folder.svg"
		}
		if dashboard.IsStarred {
			iconFile = "icons/star.svg"
		}
		icon := Icon{Path: iconFile}
		item := alfredItem{
			Arg:      targetURL.String(),
			Match:    match,
			Subtitle: dashboard.FolderTitle,
			Title:    dashboard.Title,
			UID:      dashboard.UID,
			Icon:     icon,
		}
		items = append(items, item)
	}
	return items
}

func outputJSON(items []alfredItem) {
	collection := alfredCollection{
		Items: items,
	}
	jsonData, err := json.Marshal(collection)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	fmt.Println(string(jsonData))
}
