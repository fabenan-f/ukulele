package main

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func handleRecommendation(w http.ResponseWriter, req *http.Request) {
	r, err := decodeJSON(req)
	if err != nil {
		log.Fatal(err)
	}
	categorizeRecommendationResults(r)
}

func categorizeRecommendationResults(res results) {
	var batPro, sam, resDet, resDetNA bool
	for _, result := range res.Results {
		switch metric := sliceCheckId(result.CheckId); metric {
		case "batch processing":
			batPro = true
		case "sampling":
			sam = true
		case "resource detector":
			if result.Extra.Message == "not part of library" {
				resDetNA = true
			} else {
				resDet = true
			}
		default:
			log.WithFields(logrus.Fields{
				"unknown": result.Extra.Message,
			}).Error("metric is not defined")
		}
	}

	fmt.Print("################ RECOMMENDATION ################\n")

	// invoke metric-specific analysis
	analyzeBatPro(batPro)
	analyzeSam(sam)
	analyzeResDet(resDet, resDetNA)
}

// metric: UsesBatchProcessing
func analyzeBatPro(batPro bool) {
	if !batPro {
		log.Warn("Send spans in batches to avoid network overhead")
	}
}

// metric: UsesSampling
func analyzeSam(sam bool) {
	if !sam {
		log.Warn("Use sampling to avoid performance issues")
	}
}

// metric: UsesResourceDetector
func analyzeResDet(resDet bool, resDetNA bool) {
	if !resDet && !resDetNA {
		log.Warn("Scan the system for relevant data via a resource detector")
	} else if resDetNA {
		log.Warn("Language doesn't provide proper recourse detectors")
	}
}
