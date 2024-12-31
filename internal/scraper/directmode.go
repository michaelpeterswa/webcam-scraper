package scraper

import (
	"fmt"
	"log/slog"
	"net/url"
)

func (s *Scraper) Direct(sp ScrapePoint) {
	switch sp.PageType {
	case PageTypeDirect:
		webcamURL, err := url.Parse(sp.Address)
		if err != nil {
			slog.Error("could not parse webcam url", slog.String("url", sp.Address))
		}

		slog.Info("found webcam image", slog.String("address", sp.Address), slog.String("url", webcamURL.String()))

		err = s.DownloadAndSaveImage(fmt.Sprintf(sp.FileFormatString, s.Config.OutputImageDirectory), webcamURL.String())
		if err != nil {
			slog.Error("could not download and save image", slog.String("error", err.Error()))
			return
		}

		slog.Info("downloaded and saved image", slog.String("domain", sp.Address), slog.String("url", webcamURL.String()))
	default:
		slog.Error("unknown page type", slog.String("domain", sp.Address))
	}

}
