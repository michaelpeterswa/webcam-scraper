package scraper

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"alpineworks.io/wsdot"
	"alpineworks.io/wsdot/cameras"
	"github.com/PuerkitoBio/goquery"
	"github.com/michaelpeterswa/go-start/internal/image"
)

type PageType int
type CollectionType int

const (
	PageTypeWeatherBug PageType = iota
	PageTypeWSDOT

	CollectionTypeScrape CollectionType = iota
	CollectionTypeAPI
)

type ScrapePoint struct {
	Address          string
	PageType         PageType
	FileFormatString string
	Mode             CollectionType
}

type ScraperConfig struct {
	UserAgent            string
	ScrapePoints         []ScrapePoint
	OutputImageDirectory string
}

func NewScraperConfig(ua string, sp []ScrapePoint, outputImageDirectory string) *ScraperConfig {
	return &ScraperConfig{
		UserAgent:            ua,
		ScrapePoints:         sp,
		OutputImageDirectory: outputImageDirectory,
	}
}

type Scraper struct {
	Client        *http.Client
	Config        *ScraperConfig
	CamerasClient *cameras.CamerasClient
}

type ScraperOption func(*Scraper)

func NewScraper(ctx context.Context, so ...ScraperOption) *Scraper {
	scraper := &Scraper{}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	scraper.Client = client

	for _, o := range so {
		o(scraper)
	}

	return scraper
}

func WithConfig(c *ScraperConfig) ScraperOption {
	return func(s *Scraper) {
		s.Config = c
	}
}

func WithWSDOTClient(wc *wsdot.WSDOTClient) ScraperOption {
	if wc == nil {
		return func(s *Scraper) {
			s.CamerasClient = nil
		}
	}

	camerasClient, err := cameras.NewCamerasClient(wc)
	if err != nil {
		slog.Error("could not create cameras client", slog.String("error", err.Error()))

		return func(s *Scraper) {
			s.CamerasClient = nil
		}
	}

	return func(s *Scraper) {
		s.CamerasClient = camerasClient
	}
}

func (s *Scraper) Visit(url string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", s.Config.UserAgent)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	slog.Debug("visited domain", slog.String("url", url))

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return doc, nil
}

func (s *Scraper) Run() {
	for _, scrapePoint := range s.Config.ScrapePoints {
		switch scrapePoint.Mode {
		case CollectionTypeScrape:
			s.Scrape(scrapePoint)
		case CollectionTypeAPI:
			s.API(scrapePoint)

		}
	}
}

func (s *Scraper) DownloadAndSaveImage(filename string, url string) error {
	// take directory minus filename
	directoryPath := filepath.Dir(filename)

	// Ensure the directory exists
	err := os.MkdirAll(directoryPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("User-Agent", s.Config.UserAgent)

	resp, err := s.Client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch image: %s", resp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	pngBytes, err := image.ToPNG(imageBytes)
	if err != nil {
		return fmt.Errorf("failed to convert image to png: %w", err)
	}

	_, err = file.Write(pngBytes)
	if err != nil {
		return fmt.Errorf("failed to write image to file: %w", err)
	}

	return nil
}
