package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

func GetEnvDefault(key, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

func invoke_ws_list(url_list string, req *http.Request) int {

	VERSION := GetEnvDefault("VERSION", "")
	APP := GetEnvDefault("APP", "")

	//Tracing
	closer, err := initializeTracer()
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	tracer := opentracing.GlobalTracer()

	// initialize tracer
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := tracer.StartSpan("generic_service-"+APP+"-"+VERSION, ext.RPCServerOption(spanCtx))
	defer serverSpan.Finish()

	total := 0
	urls := strings.Split(url_list, ",")

	for i, s := range urls {
		fmt.Println(i, s)

		// create child span
		clientSpan := tracer.StartSpan("generic_service_"+APP+"_"+VERSION+"_event", opentracing.ChildOf(serverSpan.Context()))
		defer clientSpan.Finish()
		tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

		res, err := http.Get(s)

		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", body)
	}

	return total + 1
}

func backgroundTask() {
	SCHED_CALL_INTERVAL := GetEnvDefault("SCHED_CALL_INTERVAL", "5")
	duration, _ := time.ParseDuration(SCHED_CALL_INTERVAL + "s")
	ticker := time.NewTicker(duration)
	for _ = range ticker.C {
		fmt.Println("invoking every " + SCHED_CALL_INTERVAL + "s")
		// Get enviroment
		SCHED_CALL_URL_LST := GetEnvDefault("SCHED_CALL_URL_LST", "")

		req, _ := http.NewRequest("GET", "http://localhost:8000", nil)

		if SCHED_CALL_URL_LST != "" {
			invoke_ws_list(SCHED_CALL_URL_LST, req)
		}
	}
}

type Message struct {
	Name        string
	Description string
	When        string
	App         string
	Version     string
}

func initializeTracer() (io.Closer, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: "user",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger

	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
	)

	VERSION := GetEnvDefault("VERSION", "")
	APP := GetEnvDefault("APP", "")

	testSpan := tracer.StartSpan("generic-call")
	testSpan.LogKV("app", APP)
	testSpan.SetTag("version", VERSION)

	// do more thing
	testSpan.Finish()
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	opentracing.SetGlobalTracer(tracer)
	return closer, err
}

func main() {
	// Healthz
	healthHandler := func(w http.ResponseWriter, req *http.Request) {
		// current time
		t := time.Now()

		VERSION := GetEnvDefault("VERSION", "")
		APP := GetEnvDefault("APP", "")

		message := Message{"Status", "health", t.Format("2006-01-02 15:04:05"), APP, VERSION}

		js, err := json.Marshal(message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

	// Split
	splitHandler := func(w http.ResponseWriter, req *http.Request) {
		// Get enviroment
		SPLIT_CALL_URL_LST := GetEnvDefault("SPLIT_CALL_URL_LST", "")
		VERSION := GetEnvDefault("VERSION", "")
		APP := GetEnvDefault("APP", "")

		invoke_ws_list(SPLIT_CALL_URL_LST, req)

		t := time.Now()

		message := Message{"split", SPLIT_CALL_URL_LST, t.Format("2006-01-02 15:04:05"), APP, VERSION}

		js, err := json.Marshal(message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

	// Response code
	responseHandler := func(w http.ResponseWriter, req *http.Request) {
		code, _ := strconv.Atoi(req.URL.Query().Get("code"))
		delay := req.URL.Query().Get("delay")

		t := time.Now()
		when := t.Format("2006-01-02 15:04:05")
		wait, _ := time.ParseDuration(delay + "s")

		time.Sleep(wait)

		nt := time.Now()
		now := nt.Format("2006-01-02 15:04:05")

		http.Error(w, fmt.Sprintf("At %v this request asks for %v and waits for %v(s) and now is %v.", when, code, delay, now), code)
	}

	go backgroundTask()

	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/s", splitHandler)
	http.HandleFunc("/r", responseHandler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
