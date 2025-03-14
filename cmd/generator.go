package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"example.com/analytics_api/internal/events"
	"example.com/analytics_api/pkg/config"
	log "github.com/sirupsen/logrus"
)

func generator(ctx context.Context, cr config.IReader) {
	generatorCr, err := config.Sub(cr, "generator")
	if err != nil {
		log.Error(err)
		return
	}
	rps, err := config.Get[int](generatorCr, "rps")
	if err != nil {
		log.Error(err)
		return
	}
	if rps <= 0 {
		log.Error("generator disabled")
		return
	}

	addr, err := config.Get[string](generatorCr, "addr")
	if err != nil {
		log.Error(err)
		return
	}

	interval := time.Second / time.Duration(rps)
	log.Infof("start ticker each %s", interval)
	t := time.NewTicker(interval)
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			jsonStr, err := json.Marshal(events.ApiEvent{
				Dt:     time.Now().Add(-24 * 365 * time.Hour),
				Event:  "generator",
				UserId: strconv.Itoa(i),
			})
			if err != nil {
				log.Error(err)
			}
			req, err := http.NewRequest("POST", addr, bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			if err != nil {
				log.Error(err)
			}
			client := &http.Client{}
			resp, err := client.Do(req)

			if err != nil {
				log.Error(err)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != 200 {
				log.Errorf("response error %v: %s", resp.Status, body)
			}
		}
		i++
	}
}
