package scraper

import (
	"fmt"
	"log/slog"
	"strconv"
)

func (s *Scraper) API(sp ScrapePoint) {
	switch sp.PageType {
	case PageTypeWSDOT:
		if s.CamerasClient == nil {
			slog.Error("wsdot cameras client is nil")
			return
		}

		id, err := strconv.Atoi(sp.Address)
		if err != nil {
			slog.Error("could not convert address to int", slog.String("address", sp.Address))
			return
		}

		image, err := s.CamerasClient.GetCamera(id)
		if err != nil {
			slog.Error("could not get camera", slog.String("error", err.Error()))
			return
		}

		err = s.DownloadAndSaveImage(fmt.Sprintf(sp.FileFormatString, s.Config.OutputImageDirectory), image.ImageURL)
		if err != nil {
			slog.Error("could not download and save image", slog.String("error", err.Error()))
			return
		}

		slog.Info("downloaded and saved image", slog.String("url", image.ImageURL))
	default:
		slog.Error("unknown page type", slog.String("domain", sp.Address))
	}
}
