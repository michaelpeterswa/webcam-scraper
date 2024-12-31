package scraper

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (s *Scraper) Scrape(sp ScrapePoint) {
	document, err := s.Visit(sp.Address)
	if err != nil {
		slog.Error("could not visit domain", slog.String("domain", sp.Address), slog.String("error", err.Error()))
	}

	var webcamURL *url.URL

	switch sp.PageType {
	case PageTypeWeatherBug:
		document.Find("img").Each(func(i int, s *goquery.Selection) {
			src, exists := s.Attr("src")
			if !exists {
				return
			}
			if strings.Contains(src, "cameras-cam.cdn.weatherbug.net") {
				url, err := url.Parse(src)
				if err != nil {
					slog.Error("could not parse webcam url", slog.String("url", src))
				}

				webcamURL = url
			}
		})
	case PageTypeSunMountainLodge:
		document.Find("img").Each(func(i int, s *goquery.Selection) {
			src, exists := s.Attr("src")
			if !exists {
				return
			}
			if strings.Contains(src, "smlcam.jpg") {
				url, err := url.Parse(src)
				if err != nil {
					slog.Error("could not parse webcam url", slog.String("url", src))
				}

				webcamURL = url
			}
		})
	default:
		slog.Error("unknown page type", slog.String("domain", sp.Address))
		return
	}

	if webcamURL == nil {
		slog.Error("could not find webcam image", slog.String("domain", sp.Address))
		return
	}

	slog.Info("found webcam image", slog.String("address", sp.Address), slog.String("url", webcamURL.String()))

	err = s.DownloadAndSaveImage(fmt.Sprintf(sp.FileFormatString, s.Config.OutputImageDirectory), webcamURL.String())
	if err != nil {
		slog.Error("could not download and save image", slog.String("error", err.Error()))
		return
	}

	slog.Info("downloaded and saved image", slog.String("domain", sp.Address), slog.String("url", webcamURL.String()))

}
