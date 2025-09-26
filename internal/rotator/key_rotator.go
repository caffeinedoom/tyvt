package rotator

import (
	"sync"
	"time"
)

type KeyRotator struct {
	mu               sync.RWMutex
	keys             []string
	currentIndex     int
	rotationInterval time.Duration
	lastRotation     time.Time
	started          bool
	stopChan         chan struct{}
}

func NewKeyRotator(keys []string, rotationInterval time.Duration) *KeyRotator {
	kr := &KeyRotator{
		keys:             keys,
		currentIndex:     0,
		rotationInterval: rotationInterval,
		lastRotation:     time.Now(),
		stopChan:         make(chan struct{}),
	}

	if len(keys) > 1 {
		go kr.autoRotate()
	}

	return kr
}

func (kr *KeyRotator) CurrentKey() string {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	if len(kr.keys) == 0 {
		return ""
	}

	return kr.keys[kr.currentIndex]
}

func (kr *KeyRotator) RotateKey() string {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	if len(kr.keys) <= 1 {
		return kr.CurrentKey()
	}

	kr.currentIndex = (kr.currentIndex + 1) % len(kr.keys)
	kr.lastRotation = time.Now()

	return kr.keys[kr.currentIndex]
}

func (kr *KeyRotator) GetKeyCount() int {
	kr.mu.RLock()
	defer kr.mu.RUnlock()
	return len(kr.keys)
}

func (kr *KeyRotator) GetCurrentIndex() int {
	kr.mu.RLock()
	defer kr.mu.RUnlock()
	return kr.currentIndex
}

func (kr *KeyRotator) Stop() {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	if kr.started {
		close(kr.stopChan)
		kr.started = false
	}
}

func (kr *KeyRotator) autoRotate() {
	kr.mu.Lock()
	kr.started = true
	kr.mu.Unlock()

	ticker := time.NewTicker(kr.rotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			kr.RotateKey()
		case <-kr.stopChan:
			return
		}
	}
}