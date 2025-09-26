package rotator

import (
	"testing"
	"time"
)

func TestKeyRotator_SingleKey(t *testing.T) {
	keys := []string{"key1"}
	rotator := NewKeyRotator(keys, time.Second)

	if rotator.CurrentKey() != "key1" {
		t.Errorf("Expected key1, got %s", rotator.CurrentKey())
	}

	if rotator.GetKeyCount() != 1 {
		t.Errorf("Expected 1 key, got %d", rotator.GetKeyCount())
	}
}

func TestKeyRotator_MultipleKeys(t *testing.T) {
	keys := []string{"key1", "key2", "key3"}
	rotator := NewKeyRotator(keys, time.Second)

	if rotator.CurrentKey() != "key1" {
		t.Errorf("Expected key1, got %s", rotator.CurrentKey())
	}

	rotatedKey := rotator.RotateKey()
	if rotatedKey != "key2" {
		t.Errorf("Expected key2 after rotation, got %s", rotatedKey)
	}

	rotatedKey = rotator.RotateKey()
	if rotatedKey != "key3" {
		t.Errorf("Expected key3 after rotation, got %s", rotatedKey)
	}

	rotatedKey = rotator.RotateKey()
	if rotatedKey != "key1" {
		t.Errorf("Expected key1 after wrap-around, got %s", rotatedKey)
	}
}

func TestKeyRotator_EmptyKeys(t *testing.T) {
	keys := []string{}
	rotator := NewKeyRotator(keys, time.Second)

	if rotator.CurrentKey() != "" {
		t.Errorf("Expected empty string for no keys, got %s", rotator.CurrentKey())
	}
}

func TestKeyRotator_AutoRotation(t *testing.T) {
	keys := []string{"key1", "key2"}
	rotator := NewKeyRotator(keys, 100*time.Millisecond)

	initialKey := rotator.CurrentKey()

	time.Sleep(150 * time.Millisecond)

	currentKey := rotator.CurrentKey()
	if currentKey == initialKey {
		t.Errorf("Expected key to rotate automatically, but it remained %s", currentKey)
	}

	rotator.Stop()
}