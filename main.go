// Copyright 2019 OpenTelemetry Authors
// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

//
//func initTracer() func() {
//	projectID := os.Getenv("PROJECT_ID")
//
//	// Create Google Cloud Trace exporter to be able to retrieve
//	// the collected spans.
//	exporter, err := cloudtrace.New(cloudtrace.WithProjectID(projectID))
//	if err != nil {
//		log.Fatal(err)
//	}
//	tp := sdktrace.NewTracerProvider(
//		// For this example code we use sdktrace.AlwaysSample sampler to sample all traces.
//		// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
//		sdktrace.WithSampler(sdktrace.AlwaysSample()),
//		sdktrace.WithBatcher(exporter))
//
//	otel.SetTracerProvider(tp)
//	return func() { tp.Shutdown(context.Background()) }
//}
//
//func installPropagators() {
//	otel.SetTextMapPropagator(
//		propagation.NewCompositeTextMapPropagator(
//			// Putting the CloudTraceOneWayPropagator first means the TraceContext propagator
//			// takes precedence if both the traceparent and the XCTC headers exist.
//			gcppropagator.CloudTraceOneWayPropagator{},
//			propagation.TraceContext{},
//			propagation.Baggage{},
//		))
//}
//
//func main() {
//	installPropagators()
//	shutdown := initTracer()
//	defer shutdown()
//
//	helloHandler := func(w http.ResponseWriter, req *http.Request) {
//		ctx := req.Context()
//		span := trace.SpanFromContext(ctx)
//		span.SetAttributes(attribute.String("server", "handling this..."))
//
//		_, _ = io.WriteString(w, "Hello, world!\n")
//	}
//	otelHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "Hello")
//	http.Handle("/", otelHandler)
//	err := http.ListenAndServe(":8080", nil)
//	if err != nil {
//		panic(err)
//	}
//}

func main() {
	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	// By default, traces will be sampled relatively rarely. To change the
	// sampling frequency for your entire program, call ApplyConfig. Use a
	// ProbabilitySampler to sample a subset of traces, or use AlwaysSample to
	// collect a trace on every run.
	//
	// Be careful about using trace.AlwaysSample in a production application
	// with significant traffic: a new trace will be started and exported for
	// every request.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	client := &http.Client{
		Transport: &ochttp.Transport{
			// Use Google Cloud propagation format.
			Propagation: &propagation.HTTPFormat{},
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest("GET", "https://www.google.com", nil)

		// The trace ID from the incoming request will be
		// propagated to the outgoing request.
		req = req.WithContext(r.Context())

		// The outgoing request will be traced with r's trace ID.
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		// Because we don't read the resp.Body, need to manually call Close().
		resp.Body.Close()
	})
	http.Handle("/test", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)

	// Use an ochttp.Handler in order to instrument OpenCensus for incoming
	// requests.
	httpHandler := &ochttp.Handler{
		// Use the Google Cloud propagation format.
		Propagation: &propagation.HTTPFormat{},
	}
	if err := http.ListenAndServe(":"+port, httpHandler); err != nil {
		log.Fatal(err)
	}
}
