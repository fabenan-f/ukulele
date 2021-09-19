package main

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func handleCoverage(w http.ResponseWriter, req *http.Request) {
	r, err := decodeJSON(req)
	if err != nil {
		log.Fatal(err)
	}
	categorizeCoverageResults(r)
}

func categorizeCoverageResults(res results) {
	spans := make([]result, 0, 10)
	funcs := make([]result, 0, 10)
	attrFuncs := make([]result, 0, 10)
	errHdl := make([]result, 0, 10)
	errRcd := make([]result, 0, 10)
	for _, result := range res.Results {
		switch metric := sliceCheckId(result.CheckId); metric {
		case "spans service":
			spans = append(spans, result)
		case "function counter":
			funcs = append(funcs, result)
		case "set attribute":
			attrFuncs = append(attrFuncs, result)
		case "error handlings counter":
			errHdl = append(errHdl, result)
		case "record error":
			errRcd = append(errRcd, result)
		default:
			log.WithFields(logrus.Fields{
				"unknown": result.Extra.Message,
			}).Error("metric is not defined")
		}
	}

	fmt.Print("################ COVERAGE ################\n")

	// invoke metric-specific analysis
	analyzeInstrumentedServices()
	analyzeSpans(spans)
	analyzeAttrCov(funcs, attrFuncs)
	analyzeErrCov(errHdl, errRcd)

}

// metric: InstrumentedServices
func analyzeInstrumentedServices() {
	log.WithFields(logrus.Fields{
		"instrumented": len(checkedSrvs),
		"total":        totCheckedSrvs,
	}).Info("Current state of instrumented services")

	// optional: names of already instrumented services
	// for k := range checkedServices {
	// 	log.Info(k)
	// }
}

// metric: SpansPerService
func analyzeSpans(spans []result) {
	log.WithFields(logrus.Fields{
		"amount":  len(spans),
		"service": curCheckedSrv,
	}).Info("Additional spans within service")
}

// metric: AttributeFunctionRatio
func analyzeAttrCov(funcs []result, attrFuncs []result) {
	var uniqueAttrFuncs int
	checkedFuncs := map[string]bool{}
	for _, result := range attrFuncs {
		if _, ok := checkedFuncs[result.Extra.Metavars.Func.AbstractContent]; !ok {
			checkedFuncs[result.Extra.Metavars.Func.AbstractContent] = true
			uniqueAttrFuncs++
		}
	}
	log.WithFields(logrus.Fields{
		"count": uniqueAttrFuncs,
		"total": len(funcs),
	}).Info("Functions with set attribute(s)")
}

// metric: ErrorCoverage
func analyzeErrCov(errHdl []result, errRcd []result) {
	log.WithFields(logrus.Fields{
		"recorded": len(errRcd),
		"total":    len(errHdl),
	}).Info("Errors recorded within service")
}
