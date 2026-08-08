package testutil

import "github.com/google/uuid"

// RandomUUID returns a new random UUID — convenience for tests.
func RandomUUID() uuid.UUID {
	return uuid.New()
}

// UniqueEmail returns an email guaranteed to be unique per call.
func UniqueEmail() string {
	return "test_" + uuid.New().String()[:8] + "@example.com"
}
