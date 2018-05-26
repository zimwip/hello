package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zimwip/hello/config"
	"github.com/zimwip/hello/router"

	"github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {

	var wait time.Duration
	// Parse cmd line value
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	staticDir := flag.String("dir", "./public", "Static file to server")
	appdashPort := flag.Int("appdash.port", 8700, "Run appdash locally on this port.")
	flag.Parse()

	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	cfg.OutputPaths = config.GetStringSlice("app.log.out")
	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	var tracer opentracing.Tracer
	// Would it make sense to embed Appdash?
	addr := startAppdashServer(*appdashPort)
	tracer = appdashot.NewTracer(appdash.NewRemoteCollector(addr))

	opentracing.InitGlobalTracer(tracer)

	sa := new(SocketServer)
	sa.Setup(":1234")
	go sa.Serve()

	srv := router.NewServer(":"+config.GetString("app.secured_port"), ":"+config.GetString("app.port"), *staticDir, log)

	// Création d’une variable pour l’interception du signal de fin de programme
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	// Block until we receive our signal.
	<-c
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info("shutting down")
	os.Exit(0)
}
