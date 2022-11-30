package ws

import (
	"sync"
	"time"
)

func (w *wsserver) checkTTLofRecords() {
	reversed := true
	lowerBound := time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC)
	checkTimeOfSessions := time.Second * 10 // routine time
	wg := &sync.WaitGroup{}

	ticker := time.NewTicker(checkTimeOfSessions)
	defer ticker.Stop()
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		wg.Done()
		for range ticker.C {
			upperBound := time.Now()
			iterCh, err := w.webSocketMapTTL.BoundedIterCh(reversed, lowerBound, upperBound)
			if err != nil {
				continue
			}

			for rec := range iterCh.Records() {
				w.deleteClient(rec.Key.(string))
			}

			iterCh.Close()
		}
	}(wg)

	wg.Wait()
}
