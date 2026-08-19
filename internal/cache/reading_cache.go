package cache

import (
	"sync"
	"time"

	"mine-vent/internal/model"
)

type ReadingCache struct {
	mu        sync.RWMutex
	readings  map[string]*model.Reading
	sensorMap map[string]string
	updatedAt time.Time
}

func NewReadingCache() *ReadingCache {
	return &ReadingCache{
		readings:  make(map[string]*model.Reading),
		sensorMap: make(map[string]string),
	}
}

func (rc *ReadingCache) Update(sensorID string, reading *model.Reading) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	rc.readings[sensorID] = reading
	rc.sensorMap[reading.ID] = sensorID
	rc.updatedAt = time.Now()
}

func (rc *ReadingCache) Get(sensorID string) (*model.Reading, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	r, ok := rc.readings[sensorID]
	return r, ok
}

func (rc *ReadingCache) GetAll() map[string]*model.Reading {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	result := make(map[string]*model.Reading, len(rc.readings))
	for k, v := range rc.readings {
		result[k] = v
	}
	return result
}

func (rc *ReadingCache) GetByArea(sensorIDs []string) []*model.Reading {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	var result []*model.Reading
	for _, id := range sensorIDs {
		if r, ok := rc.readings[id]; ok {
			result = append(result, r)
		}
	}
	return result
}

func (rc *ReadingCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.readings = make(map[string]*model.Reading)
	rc.sensorMap = make(map[string]string)
	rc.updatedAt = time.Time{}
}

func (rc *ReadingCache) UpdatedAt() time.Time {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.updatedAt
}

func (rc *ReadingCache) Count() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return len(rc.readings)
}

func (rc *ReadingCache) Remove(sensorID string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if r, ok := rc.readings[sensorID]; ok {
		delete(rc.sensorMap, r.ID)
	}
	delete(rc.readings, sensorID)
}

func (rc *ReadingCache) Snapshot() map[string]float64 {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	result := make(map[string]float64, len(rc.readings))
	for k, v := range rc.readings {
		result[k] = v.Value
	}
	return result
}
