package main

import (
	"encoding/json"
	"fmt"
	"github.com/leejones/netrc"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

var logLevel = os.Getenv("LOG_LEVEL")

var (
	InfoLog  *log.Logger
	DebugLog *log.Logger
	ErrorLog *log.Logger
)

func init() {
	logFileName := os.Getenv("LOG_FILE")
	if logFileName == "" {
		logFileName = "dashboards.log"
	}
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLog = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLog = log.New(logFile, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func logger(l string) {
	if logLevel == "DEBUG" {
		DebugLog.Println(l)
	} else {
		InfoLog.Println(l)
	}
}

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
	// {
	// 	"id": 48,
	// 	"uid": "24Xy_QsZz",
	// 	"title": "(C1 2020) Selective Bulk Edit",
	// 	"uri": "db/c1-2020-selective-bulk-edit",
	// 	"url": "/d/24Xy_QsZz/c1-2020-selective-bulk-edit",
	// 	"slug": "",
	// 	"type": "dash-db",
	// 	"tags": [],
	// 	"isStarred": false
	// }
	UID         string `json:"uid"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Type        string `json:"type"`
	IsStarred   bool   `json:"isStarred"`
	FolderTitle string `json:"folderTitle"`
}

func addAuth(req *http.Request) {
	apiToken := os.Getenv("GRAFANA_API_TOKEN")
	grafanaUser := os.Getenv("GRAFANA_BASIC_AUTH_USER")
	grafanaPassword := os.Getenv("GRAFANA_BASIC_AUTH_PASSWORD")

	if apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+apiToken)
	} else {
		if grafanaUser == "" || grafanaPassword == "" {
			ErrorLog.Printf("load credentials: ENV vars not set: GRAFANA_BASIC_AUTH_USER, GRAFANA_BASIC_AUTH_PASSWORD")
			basicAuth, err := netrc.Get(req.Host)
			if err != nil {
				ErrorLog.Printf("load credentials: unable to load from netrc: %v\n", err)
			} else {
				ErrorLog.Printf("load credentials: found credentials in netrc")
				grafanaUser = basicAuth.Username
				grafanaPassword = basicAuth.Password
			}
		}
		req.SetBasicAuth(grafanaUser, grafanaPassword)
	}
}

func main() {
	grafanaHost := os.Getenv("GRAFANA_HOST")
	query := strings.TrimSpace(os.Args[1])
	apiURL, err := url.Parse(grafanaHost)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	apiURL.Path = path.Join(apiURL.Path, "api/search")

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		ErrorLog.Print(err)
		os.Exit(1)
	}
	addAuth(req)
	if query != "" {
		q := req.URL.Query()
		q.Add("query", query)
		req.URL.RawQuery = q.Encode()
	}
	logger(fmt.Sprintf("Requesting: %s", req.URL.String()))
	resp, err := httpClient.Do(req)
	if err != nil {
		ErrorLog.Print(err)
		os.Exit(1)
	}
	logger(fmt.Sprintf("Response Status: %s", resp.Status))

	if resp.StatusCode != http.StatusOK {
		ErrorLog.Println("ERROR: HTTP Response:", resp.StatusCode)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorLog.Println("ERROR:", err)
		os.Exit(1)
	}

	var dashboards []dashboard
	err = json.Unmarshal(body, &dashboards)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	var items []alfredItem
	for _, dashboard := range dashboards {
		targetURL, err := url.Parse(grafanaHost)
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
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
	collection := alfredCollection{
		Items: items,
	}
	jsonData, _ := json.Marshal(collection)
	fmt.Println(string(jsonData))
}
