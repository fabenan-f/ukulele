package main

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func newRun() {
	fmt.Print(
		"#######################################\n",
		"#                                     #\n",
		"#               NEW RUN               #\n",
		"#                                     #\n",
		"#######################################\n",
	)
}

func handleSetup(w http.ResponseWriter, req *http.Request) {
	newRun()
	r, err := decodeJSON(req)
	if err != nil {
		log.Fatal(err)
	}
	categorizeSetupResults(r)
}

func categorizeSetupResults(res results) {
	genConfig := make([]result, 0, 3)
	wrpFuncs := make([]result, 0, 4)
	var rsc, ctxPro, shd result
	for _, result := range res.Results {
		switch metric := sliceCheckId(result.CheckId); metric {
		case "general configuration":
			genConfig = append(genConfig, result)
		case "resources":
			rsc = result
		case "context propagation":
			ctxPro = result
		case "wrapped functions":
			wrpFuncs = append(wrpFuncs, result)
		case "shutdown":
			shd = result
		default:
			log.WithFields(logrus.Fields{
				"metric": metric,
			}).Error("metric is not defined")
		}
	}
	fmt.Print("################ SETUP ################\n")

	// invoke metric-specific analysis
	analyzeGenConfig(genConfig)
	analyzeRsc(rsc)
	analyzeCtxPro(ctxPro)
	analyzeWrpFuncs(wrpFuncs)
	analyzeShd(shd)
}

// metric: HasGeneralConfiguration
func analyzeGenConfig(genConfig []result) {
	if len(genConfig) > 3 {
		log.WithFields(logrus.Fields{
			"expected": 3,
		}).Warn("you initialized more general operations than needed")
	}
	var exporter, spanProcessor, tracerProvider bool
	for _, result := range genConfig {
		switch result.Extra.Message {
		case "exporter":
			exporter = true
			log.Info("exporter is correctly initialized")
		case "span processor":
			spanProcessor = true
			log.Info("span processor is correctly initialized")
		case "tracer provider":
			tracerProvider = true
			log.Info("tracer provider is correctly initialized")
		default:
			log.WithFields(logrus.Fields{
				"unknown": result.Extra.Message,
			}).Error("general configuration is not defined")
		}
	}
	if !exporter {
		log.Error("exporter is not initialized")
	}
	if !spanProcessor {
		log.Error("span processor is not initialized")
	}
	if !tracerProvider {
		log.Error("tracer provider is not initialized")
	}
}

// metric: HasResources
func analyzeRsc(rsc result) {
	if rsc.Extra.Message == "resource creation" {
		log.Info("(basic) resource information is defined")
	} else {
		log.Error("no resource information is defined")
	}
}

// metric: EnablesContextPropagation
func analyzeCtxPro(ctxPro result) {
	if ctxPro.Extra.Message == "propagation requirements" {
		log.Info("context propagation is configured")
	} else {
		log.Error("missing context propagation setup")
	}
}

// metric: WrapsEndpoints
func analyzeWrpFuncs(wrpFuncs []result) {
	var httpApi, httpInstrumented, grpcApi, grpcInstrumented bool
	for _, result := range wrpFuncs {
		switch result.Extra.Message {
		case "http server check":
			httpApi = true
		case "http server instrumentation":
			httpInstrumented = true
		case "grpc server check":
			grpcApi = true
		case "grpc stream server interceptor":
			grpcInstrumented = true
		}
	}

	if httpApi && httpInstrumented {
		log.Info("http server endpoints are instrumented")
	} else if httpApi && !httpInstrumented {
		log.Error("http server endpoints are not instrumented")
	}

	if grpcApi && grpcInstrumented {
		log.Info("grpc server endpoints are instrumented")
	} else if grpcApi && !grpcInstrumented {
		log.Error("grpc server endponts are not instrumented")
	}
}

// metric: ShutdownsTracerProvider
func analyzeShd(shd result) {
	if shd.Extra.Message == "shutdown config" {
		log.Info("tracer provider will be shutdowned properly")
	} else if shd.Extra.Message == "cannot be configured" {
		log.Warn("configuration of shutdown cannot be checked in this language")
	} else {
		log.Error("shutdown of tracer provider is not configured")
	}
}
