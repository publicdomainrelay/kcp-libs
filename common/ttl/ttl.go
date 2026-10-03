package ttl

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Effective(override, fallback *int64) *int64 {
	seconds := fallback
	if override != nil {
		seconds = override
	}
	if seconds == nil || *seconds < 0 {
		return nil
	}
	value := *seconds
	return &value
}

func Expiry(completion *metav1.Time, seconds *int64) (time.Time, bool) {
	if completion == nil || seconds == nil || *seconds < 0 {
		return time.Time{}, false
	}
	return completion.Add(time.Duration(*seconds) * time.Second), true
}

func Expired(completion *metav1.Time, seconds *int64, now time.Time) (expired bool, after time.Duration, known bool) {
	expiry, ok := Expiry(completion, seconds)
	if !ok {
		return false, 0, false
	}
	if !now.Before(expiry) {
		return true, 0, true
	}
	return false, expiry.Sub(now), true
}

func Deadline(start *metav1.Time, seconds *int64, now time.Time) bool {
	if start == nil || seconds == nil || *seconds < 0 {
		return false
	}
	return !now.Before(start.Add(time.Duration(*seconds) * time.Second))
}
