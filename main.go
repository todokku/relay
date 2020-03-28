package main

import (
	"io"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

const (
	permissions = 51264
	channelID   = "ops"
)

func main() {
	InitLogging()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatalf("DISCORD_TOKEN is empty")
	}

	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Infof("Starting up on http://localhost:%s", port)

	if os.Getenv("ENABLE_STACKDRIVER") != "" {
		labels := &stackdriver.Labels{}
		labels.Set("app", "relay", "The name of the current app.")
		sd, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID:               "icco-cloud",
			MonitoredResource:       monitoredresource.Autodetect(),
			DefaultMonitoringLabels: labels,
			DefaultTraceAttributes:  map[string]interface{}{"app": "relay"},
		})

		if err != nil {
			log.WithError(err).Fatalf("failed to create the stackdriver exporter")
		}
		defer sd.Flush()

		view.RegisterExporter(sd)
		trace.RegisterExporter(sd)
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		})
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.WithError(err).Fatal("error creating Discord session")
	}

	if err := dg.Open(); err != nil {
		log.WithError(err).Fatal("error opening connection")
	}
	defer dg.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(LoggingMiddleware())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		conns, err := dg.UserConnections()
		if err != nil {
			log.WithError(err).Error("could not get connections")
		} else {
			log.WithFields(logrus.Fields{"connections": conns}).Debug("user connections")
		}

		w.Write([]byte("hi."))
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi."))
	})

	r.Post("/hook", func(w http.ResponseWriter, r *http.Request) {
		var message []byte
		if _, err := io.ReadFull(r.Body, message); err != nil {
			log.WithError(err).Error("could not read body")
			http.Error(w, err.Error(), 500)
			return
		}

		if err := messageCreate(dg, string(message)); err != nil {
			log.WithError(err).Error("could not send message")
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write([]byte("."))
	})

	h := &ochttp.Handler{
		Handler:     r,
		Propagation: &propagation.HTTPFormat{},
	}
	if err := view.Register([]*view.View{
		ochttp.ServerRequestCountView,
		ochttp.ServerResponseCountByStatusCode,
	}...); err != nil {
		log.WithError(err).Fatal("Failed to register ochttp views")
	}
	log.Fatal(http.ListenAndServe(":"+port, h))
}

func messageCreate(s *discordgo.Session, m string) error {
	_, err := s.ChannelMessageSend(channelID, "Pong!")
	return err
}
