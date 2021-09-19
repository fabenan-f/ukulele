package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// deserialized Semgrep JSON
type results struct {
	Results []result `json:"results"`
}

type result struct {
	CheckId string `json:"check_id"`
	Path    string `json:"path"`
	Start   place  `json:"start"`
	End     place  `json:"end"`
	Extra   extra  `json:"extra"`
}

type place struct {
	Line   int `json:"line"`
	Col    int `json:"col"`
	Offset int `json:"offset"`
}

type extra struct {
	Message   string        `json:"message"`
	Metavars  metavarOption `json:"metavars"`
	Metadata  interface{}   `json:"metadata"`
	Severity  string        `json:"severity"`
	IsIgnored bool          `json:"is_ignored"`
	Lines     string        `json:"lines"`
}

type metavarOption struct {
	Span  metavar `json:"$SPAN"`
	Func  metavar `json:"$FUNC"`
	Exp   metavar `json:"$EXP"`
	Sdk   metavar `json:"$SDK"`
	Key   metavar `json:"$KEY"`
	Value metavar `json:"$VALUE"`
}

type metavar struct {
	Start           place    `json:"start"`
	End             place    `json:"end"`
	AbstractContent string   `json:"abstract_content"`
	UniqueID        uniqueID `json:"unique_id"`
}

type uniqueID struct {
	MetavarType string `json:"type"`
	Md5sum      string `json:"md5sum"`
}

var log = logrus.New()

func decodeJSON(req *http.Request) (results, error) {
	decoder := json.NewDecoder(req.Body)
	var r results
	err := decoder.Decode(&r)
	return r, err
}

func sliceCheckId(path string) string {
	substring := strings.SplitAfter(path, ".")
	return substring[len(substring)-1]
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/setup", handleSetup).Methods(http.MethodPost)
	r.HandleFunc("/bestpractice", handleBestPractice).Methods(http.MethodPost)
	r.HandleFunc("/coverage", handleCoverage).Methods(http.MethodPost)
	r.HandleFunc("/recommendation", handleRecommendation).Methods(http.MethodPost)
	var handler http.Handler = r

	// IP address needs to be configured in order to work with GitHub Actions
	log.Fatal(http.ListenAndServe("localhost:8080", handler))
}
