package ws

import (
	"time"
)

func (w *wsserver) checkTTLofRecords() {
	reversed := true
	lowerBound := time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC)

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for range ticker.C {
		upperBound := time.Now()
		iterCh, err := w.wsmapTTL.BoundedIterCh(reversed, lowerBound, upperBound)
		if err != nil {
			continue
		}

		for rec := range iterCh.Records() {
			w.onDelClient(rec.Key.(string))
		}

		iterCh.Close()
	}

}
