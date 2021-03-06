package config

import (
	"time"
	"fmt"
)

type CleanupMethod string
const (
	PingBasedCleanup = CleanupMethod("ping based cleanup")
	TimeBasedCleanup = CleanupMethod("time based cleanup")
)



type Config struct {
	CleanupLeasesInterval  time.Duration
	CleanupMethod          CleanupMethod
	MissedPingsThreshold   int
	LeasesExpiryDuration   time.Duration
	PersistentLeases       bool
}



func NewDefaultConfig() Config {
	cleanupLeasesInterval,  _ := time.ParseDuration("60m")
	leasesExpiryDuration, _ := time.ParseDuration(fmt.Sprintf("%dh", 7 * 24))

	return Config {
		CleanupLeasesInterval: cleanupLeasesInterval,
		CleanupMethod:         PingBasedCleanup,
		MissedPingsThreshold:  5,
		LeasesExpiryDuration:  leasesExpiryDuration,
		PersistentLeases:      false,
	}
}
