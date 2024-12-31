package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"alpineworks.io/wsdot"
	"github.com/alpineworks/ootel"
	"github.com/michaelpeterswa/go-start/internal/config"
	"github.com/michaelpeterswa/go-start/internal/logging"
	"github.com/michaelpeterswa/go-start/internal/scraper"
	"github.com/robfig/cron/v3"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "error"
	}

	slogLevel, err := logging.LogLevelToSlogLevel(logLevel)
	if err != nil {
		log.Fatalf("could not convert log level: %s", err)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})))

	ctx := context.Background()

	config, err := config.NewConfig()
	if err != nil {
		slog.Error("could not create config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ootelClient := ootel.NewOotelClient(
		ootel.WithMetricConfig(
			ootel.NewMetricConfig(
				config.MetricsEnabled,
				config.MetricsPort,
			),
		),
		ootel.WithTraceConfig(
			ootel.NewTraceConfig(
				config.TracingEnabled,
				config.TracingSampleRate,
				config.TracingService,
				config.TracingVersion,
			),
		),
	)

	shutdown, err := ootelClient.Init(ctx)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = shutdown(ctx)
	}()

	scraperConfig := scraper.NewScraperConfig(
		"github.com/michaelpeterswa/webcam-scraper",
		mapToDomains(config.Domains),
		config.OutputImageDir,
	)

	var scraperOptions []scraper.ScraperOption

	if config.WSDOTAPIKey != "" {
		wc, err := wsdot.NewWSDOTClient(wsdot.WithAPIKey(config.WSDOTAPIKey), wsdot.WithHTTPClient(
			&http.Client{
				Timeout: 10 * time.Second,
			},
		))
		if err != nil {
			slog.Error("could not create WSDOT client", slog.String("error", err.Error()))
		} else {
			scraperOptions = append(scraperOptions, scraper.WithWSDOTClient(wc))
		}
	}

	scraperOptions = append(scraperOptions, scraper.WithConfig(scraperConfig))

	s := scraper.NewScraper(ctx, scraperOptions...)

	c := cron.New()
	_, err = c.AddFunc(config.CronSchedule, s.Run)
	if err != nil {
		slog.Error("could not add cron job", slog.String("error", err.Error()))
	}

	c.Start()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	go func() {
		sig := <-sigs

		slog.Info("received signal", slog.String("signal", sig.String()))
		c.Stop()

		done <- true
	}()

	slog.Info("webcam-scraper is running")
	<-done
	slog.Info("webcam-scraper is shutting down")

}

func mapToDomains(m map[string]string) []scraper.ScrapePoint {
	var d []scraper.ScrapePoint

	for k, v := range m {
		fileTypeMode := strings.Split(k, "-")
		file := fileTypeMode[0]
		pageType := fileTypeMode[1]
		mode := fileTypeMode[2]

		var translatedPageType scraper.PageType
		switch strings.ToLower(pageType) {
		case "weatherbug":
			translatedPageType = scraper.PageTypeWeatherBug
		case "wsdot":
			translatedPageType = scraper.PageTypeWSDOT
		case "direct":
			translatedPageType = scraper.PageTypeDirect
		case "sunmountainlodge":
			translatedPageType = scraper.PageTypeSunMountainLodge
		default:
			slog.Error("could not translate mode", slog.String("mode", mode))
			continue
		}

		var translatedMode scraper.CollectionType
		switch strings.ToLower(mode) {
		case "scrape":
			translatedMode = scraper.CollectionTypeScrape
		case "api":
			translatedMode = scraper.CollectionTypeAPI
		case "direct":
			translatedMode = scraper.CollectionTypeDirect
		default:
			slog.Error("could not translate page type", slog.String("pageType", pageType))
			continue
		}

		d = append(d, scraper.ScrapePoint{
			Address:          v,
			PageType:         translatedPageType,
			Mode:             translatedMode,
			FileFormatString: "%s" + fmt.Sprintf("/%s/%s/%s.png", strings.ToLower(pageType), strings.ToLower(mode), strings.ToLower(file)),
		})
	}

	return d
}
