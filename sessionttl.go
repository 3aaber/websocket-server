package ws

import (
	"os"
	"os/signal"
	"time"
)

func (w *wsserver) checkTTLofRecords() {
	reversed := true
	lowerBound := time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC)
	upperBound := time.Now()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for range ticker.C {

		iterCh, err := w.wsmapTTL.BoundedIterCh(reversed, lowerBound, upperBound)
		if err != nil {
			continue
		}
		// defer iterCh.Close()

		for rec := range iterCh.Records() {
			w.onDelClient(rec.Key.(string))
		}
	}

}
