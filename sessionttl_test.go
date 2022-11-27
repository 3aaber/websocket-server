package ws

import (
	"fmt"
	"testing"
	"time"

	"github.com/umpc/go-sortedmap"
	"github.com/umpc/go-sortedmap/asc"
)

func TestSortedMap(t *testing.T) {
	// Create an empty SortedMap with a size suggestion and a less than function:
	sm := sortedmap.New(4, asc.Time)

	// Insert example records:
	sm.Insert("Now - 3", time.Now().Add(-time.Hour*2))
	sm.Insert("Now - 2", time.Now().Add(-time.Hour*2))
	sm.Insert("Now - 1", time.Now().Add(-time.Hour*1))
	sm.Insert("Now + 1", time.Now().Add(time.Hour*1))
	sm.Insert("Now + 2", time.Now().Add(time.Hour*2))
	sm.Insert("Now + 3", time.Now().Add(time.Hour*2))

	// Set iteration options:
	reversed := true
	lowerBound := time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC)
	upperBound := time.Now()

	// Select values > lowerBound and values <= upperBound.
	// Loop through the values, in reverse order:
	iterCh, err := sm.BoundedIterCh(reversed, lowerBound, upperBound)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer iterCh.Close()

	for rec := range iterCh.Records() {
		fmt.Printf("%+v\n", rec)
	}
}
