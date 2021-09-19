package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// deserialized custom configuration JSON
type srvConfigs struct {
	TotalSrvs  int         `json:"totalAmountOfServices"`
	SrvConfigs []srvConfig `json:"serviceConfigs"`
}

type srvConfig struct {
	Service string `json:"service"`
	Config  config `json:"config"`
}

type config struct {
	Namespace       string          `json:"namespace"`
	Dependencies    []dependency    `json:"dependencies"`
	UnitsAttributes []unitAttribute `json:"unitsAttributes"`
}

type dependency struct {
	Name    string `json:"dependency"`
	Version string `json:"version"`
}

type unitAttribute struct {
	Attribute string `json:"attribute"`
	Unit      string `json:"unit"`
}

// information for coverage
var checkedSrvs map[string]bool = make(map[string]bool, 10)
var totCheckedSrvs int
var curCheckedSrv string

func handleBestPractice(w http.ResponseWriter, req *http.Request) {
	r, err := decodeJSON(req)
	if err != nil {
		log.Fatal(err)
	}
	categorizeBestPracticeResults(r)
}

func categorizeBestPracticeResults(res results) {
	var srvName string
	attr := make([]result, 0, 10)
	errStates := make([]result, 0, 5)
	spanEnds := make([]result, 0, 2)
	var semCon bool
	for _, result := range res.Results {
		switch metric := sliceCheckId(result.CheckId); metric {
		case "id for config file":
			srvName = result.Extra.Message
		case "error state":
			errStates = append(errStates, result)
		case "span end":
			spanEnds = append(spanEnds, result)
		case "semantic conventions":
			semCon = true
		case "checking attributes":
			attr = append(attr, result)
		default:
			log.WithFields(logrus.Fields{
				"unknown": result.Extra.Message,
			}).Error("metric is not defined")
		}
	}
	curCheckedSrv = srvName
	checkedSrvs[srvName] = true
	config, err := getSrvInfo(srvName)
	if err != nil {
		log.Error(err)
	}

	fmt.Print("################ BEST PRACTICE ################\n")

	// invoke metric-specific analysis
	analyzeSpanEnd(spanEnds)
	analyzeErrState(errStates)
	analyzeSemCon(semCon)
	analyzeCustConfig(attr, *config)
}

// metric: SetsErrorState
func analyzeErrState(errStates []result) {
	if len(errStates) > 0 && errStates[0].Extra.Message == "cannot be configured" {
		log.Warn("Error states cannot be checked in this language")
		return
	}

	for _, errState := range errStates {
		log.WithFields(logrus.Fields{
			// where span starts
			"line": errState.Extra.Metavars.Span.Start.Line,
		}).Error("Error state is not attached to span")
	}
}

// metric: EndsSpan
func analyzeSpanEnd(spanEnds []result) {
	if len(spanEnds) > 0 && spanEnds[0].Extra.Message == "cannot be configured" {
		log.Warn("Span ends cannot be checked in this language")
		return
	}

	for _, spanEnd := range spanEnds {
		log.WithFields(logrus.Fields{
			// where span starts
			"line": spanEnd.Extra.Metavars.Span.Start.Line,
		}).Error("Span does not properly end (consider defer)")
	}
}

// metric: UsesSemanticConventions
func analyzeSemCon(semCon bool) {
	if semCon {
		log.Info("Semantic conventions are in use")
	} else {
		log.Warn("No semanic conventions are in use")
	}
}

func analyzeCustConfig(attr []result, config config) {
	if len(attr) == 0 {
		log.Warn("There are no attributes attached")
		return
	}
	analyzePII(attr)
	analyzeNsp(attr, config.Namespace)
	anaylzeDpd(attr, config.Dependencies)
	analyzeAttrU(attr, config.UnitsAttributes)
}

// metric: ExcludesPersonalInformation
func analyzePII(attr []result) {
	// just an idea how checking for personally identifiable information (PII) could look like
	for _, a := range attr {
		if strings.Contains(a.Extra.Metavars.Key.AbstractContent, "@gmail.com") {
			log.WithFields(logrus.Fields{
				"key":  a.Extra.Metavars.Key.AbstractContent,
				"line": a.Extra.Metavars.Key.Start.Line,
			}).Warn("Attribute might contain private information")
		}
		if strings.Contains(a.Extra.Metavars.Value.AbstractContent, "@gmail.com") {
			log.WithFields(logrus.Fields{
				"value": a.Extra.Metavars.Value.AbstractContent,
				"line":  a.Extra.Metavars.Value.Start.Line,
			}).Warn("Attribute might contain private information")
		}
	}
}

// metric: UsesNamespaces
func analyzeNsp(attr []result, nsp string) {
	npcRegex, err := regexp.Compile("^\"" + nsp + ".*")
	if err != nil {
		log.Error(err)
	}
	for _, a := range attr {
		match := npcRegex.MatchString(a.Extra.Metavars.Key.AbstractContent)
		if match {
			log.WithFields(logrus.Fields{
				"line": a.Extra.Metavars.Key.Start.Line,
			}).Info("Found correct namespace")
		} else {
			log.WithFields(logrus.Fields{
				"line":              a.Extra.Metavars.Key.Start.Line,
				"correct namespace": nsp,
			}).Error("Incorrect namespace")
		}
	}
}

// metric: IncludesDependencyVersion
func anaylzeDpd(attr []result, dpds []dependency) {
	// create dependecy map
	mapDpd := make(map[string]string, len(dpds))
	for _, dpd := range dpds {
		mapDpd[dpd.Name] = dpd.Version
	}

	for _, a := range attr {
		for _, dpd := range dpds {
			if strings.Contains(a.Extra.Metavars.Key.AbstractContent, dpd.Name) && strings.Contains(a.Extra.Metavars.Value.AbstractContent, dpd.Version) {
				delete(mapDpd, dpd.Name)
				log.WithFields(logrus.Fields{
					"dependecy": dpd.Name,
				}).Info("Dependency is covered")
			}
		}
	}

	for dpd, ver := range mapDpd {
		log.WithFields(logrus.Fields{
			"dependency": dpd,
			"version":    ver,
		}).Error("Dependency is not covered or version is incorrect")
	}
}

// metric: IncludesUnits
func analyzeAttrU(attr []result, attrU []unitAttribute) {
	for _, a := range attr {
		for _, u := range attrU {
			if strings.Contains(a.Extra.Metavars.Key.AbstractContent, u.Attribute) && !strings.Contains(a.Extra.Metavars.Key.AbstractContent, u.Unit) {
				log.WithFields(logrus.Fields{
					"attribute":    u.Attribute,
					"missing unit": u.Unit,
				}).Error("Attribute should contain unit")
			}
		}
	}
}

func readFromJSONFile() []uint8 {
	df, err := os.Open("custom-config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer df.Close()
	bv, err := ioutil.ReadAll(df)
	if err != nil {
		log.Fatal(err)
	}
	return bv
}

func getSrvInfo(srv string) (*config, error) {
	bv := readFromJSONFile()
	srvConfigs := new(srvConfigs)
	json.Unmarshal(bv, &srvConfigs)

	totCheckedSrvs = srvConfigs.TotalSrvs

	for _, srvConfig := range srvConfigs.SrvConfigs {
		if srvConfig.Service == srv {
			return &srvConfig.Config, nil
		}
	}
	return nil, errors.New("configuration does not exist")
}
