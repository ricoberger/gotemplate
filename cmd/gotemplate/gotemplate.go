package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ricoberger/gotemplate/pkg/api"
	"github.com/ricoberger/gotemplate/pkg/log"
	"github.com/ricoberger/gotemplate/pkg/metrics"
	"github.com/ricoberger/gotemplate/pkg/tracer"
	"github.com/ricoberger/gotemplate/pkg/version"
	"go.uber.org/zap"

	flag "github.com/spf13/pflag"
)

const NAME = "gotemplate"

var (
	logFormat            string
	logLevel             string
	traceEnabled         bool
	traceProviderName    string
	traceProviderAddress string
	apiAddress           string
	metricsAddress       string
	showVersion          bool
)

func init() {
	defaultLogFormat := "console"
	if os.Getenv("GOTEMPLATE_LOG_FORMAT") != "" {
		defaultLogFormat = os.Getenv("GOTEMPLATE_LOG_FORMAT")
	}

	defaultLogLevel := "info"
	if os.Getenv("GOTEMPLATE_LOG_LEVEL") != "" {
		defaultLogLevel = os.Getenv("GOTEMPLATE_LOG_LEVEL")
	}

	defaultTraceProviderName := "jaeger"
	if os.Getenv("GOTEMPLATE_TRACE_PROVIDER_NAME") != "" {
		defaultTraceProviderName = os.Getenv("GOTEMPLATE_TRACE_PROVIDER_NAME")
	}

	defaultTraceProviderAddress := "http://localhost:14268/api/traces"
	if os.Getenv("GOTEMPLATE_TRACE_PROVIDER_ADDRESS") != "" {
		defaultTraceProviderAddress = os.Getenv("GOTEMPLATE_TRACE_PROVIDER_ADDRESS")
	}

	defaultMetricsAddress := ":8081"
	if os.Getenv("GOTEMPLATE_METRICS_ADDRESS") != "" {
		defaultMetricsAddress = os.Getenv("GOTEMPLATE_METRICS_ADDRESS")
	}

	defaultAPIAddress := ":8080"
	if os.Getenv("GOTEMPLATE_API_ADDRESS") != "" {
		defaultAPIAddress = os.Getenv("GOTEMPLATE_API_ADDRESS")
	}

	flag.StringVar(&logFormat, "log.format", defaultLogFormat, "Set the output format of the logs. Must be \"console\" or \"json\".")
	flag.StringVar(&logLevel, "log.level", defaultLogLevel, "Set the log level. Must be \"debug\", \"info\", \"warn\", \"error\", \"fatal\" or \"panic\".")
	flag.BoolVar(&traceEnabled, "trace.enabled", false, "Enable / disable tracing.")
	flag.StringVar(&traceProviderName, "trace.provider.name", defaultTraceProviderName, "Select the tracing provider which should be used. Must be \"jaeger\" or \"zipkin\".")
	flag.StringVar(&traceProviderAddress, "trace.provider.address", defaultTraceProviderAddress, "The address of the tracing provider.")
	flag.StringVar(&metricsAddress, "metrics.address", defaultMetricsAddress, "Set the address where the metrics server should listen on.")
	flag.StringVar(&apiAddress, "api.address", defaultAPIAddress, "Set the address where the API server should listen on,")
	flag.BoolVar(&showVersion, "version", false, "Print version information.")
}

func main() {
	// Parse our command-line flags and setup our logger and tracer.
	flag.Parse()
	log.Setup(logLevel, logFormat)

	if traceEnabled {
		err := tracer.Setup(NAME, traceProviderName, traceProviderAddress)
		if err != nil {
			log.Fatal(nil, "Could not setup tracing", zap.Error(err), zap.String("provider-name", traceProviderName), zap.String("provider-address", traceProviderAddress))
		}
	}

	// If the version flag is set we print the version information and exit the application.
	if showVersion {
		v, err := version.Print(NAME)
		if err != nil {
			log.Fatal(nil, "Failed to print version information", zap.Error(err))
		}

		fmt.Fprintln(os.Stdout, v)
		return
	}

	// Print the short form for our version information.
	log.Info(nil, "Version information", version.Info()...)
	log.Info(nil, "Build context", version.BuildContext()...)

	// Create a new API and metrics server and start them in a new Go routine.
	apiServer, err := api.New(apiAddress)
	if err != nil {
		log.Fatal(nil, "Could not create API server", zap.Error(err))
	}
	go apiServer.Start()

	metricsServer := metrics.New(metricsAddress)
	go metricsServer.Start()

	// All components should be terminated gracefully. For that we are listen for the SIGINT and SIGTERM signals and try
	// to gracefully shutdown the started servers. This ensures that established connections or tasks are not
	// interrupted.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	log.Debug(nil, "Start listining for SIGINT and SIGTERM signal")
	<-done
	log.Info(nil, "Start shutdown...")

	metricsServer.Stop()
	apiServer.Stop()

	log.Info(nil, "Shutdown is done")
}
