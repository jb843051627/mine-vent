package cache

import (
	"sync"
	"testing"

	"mine-vent/internal/model"
)

func TestMV03_ConcurrentCacheUpdateDataRace(t *testing.T) {
	rc := NewReadingCache()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			rc.Update("sensor-1", &model.Reading{
				ID: "r", SensorID: "sensor-1", Value: float64(n),
			})
		}(i)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rc.Get("sensor-1")
		}()
	}
	wg.Wait()

	r, _ := rc.Get("sensor-1")
	if r == nil {
		t.Error("expected non-nil reading after updates")
	}
}
