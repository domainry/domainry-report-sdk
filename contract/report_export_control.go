package contract

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	reportmodel "github.com/domainry/domainry-report-sdk/model"
)

const MinimumReportExportDownloadTTLSeconds int64 = 60

// ReportExportDownloadTTLSeconds resolves the business-authored typed value
// while retaining compatibility with config.download_ttl_seconds. Its only
// upper bound is Go's duration representation, not a product retention policy.
func ReportExportDownloadTTLSeconds(control reportmodel.ReportExportControlSchema, defaultSeconds int64) (int64, error) {
	typed, typedSet := control.DownloadTTLSeconds, control.DownloadTTLSeconds != 0
	legacy, legacySet, err := reportExportLegacyDownloadTTLSeconds(control.Config)
	if err != nil {
		return 0, err
	}
	if typedSet && legacySet && typed != legacy {
		return 0, fmt.Errorf("download_ttl_seconds conflicts with config.download_ttl_seconds")
	}
	seconds, set := typed, typedSet
	if !set && legacySet {
		seconds, set = legacy, true
	}
	if !set {
		seconds = defaultSeconds
	}
	const maximumDurationSeconds = int64(math.MaxInt64) / int64(1_000_000_000)
	if seconds < MinimumReportExportDownloadTTLSeconds || seconds > maximumDurationSeconds {
		return 0, fmt.Errorf("download_ttl_seconds must be an integer between %d and %d", MinimumReportExportDownloadTTLSeconds, maximumDurationSeconds)
	}
	return seconds, nil
}

func reportExportLegacyDownloadTTLSeconds(config map[string]any) (int64, bool, error) {
	raw, set := config["download_ttl_seconds"]
	if !set {
		return 0, false, nil
	}
	invalid := func() (int64, bool, error) {
		return 0, true, fmt.Errorf("config.download_ttl_seconds must be an integer")
	}
	switch value := raw.(type) {
	case int:
		return int64(value), true, nil
	case int8:
		return int64(value), true, nil
	case int16:
		return int64(value), true, nil
	case int32:
		return int64(value), true, nil
	case int64:
		return value, true, nil
	case uint:
		if uint64(value) > math.MaxInt64 {
			return invalid()
		}
		return int64(value), true, nil
	case uint8:
		return int64(value), true, nil
	case uint16:
		return int64(value), true, nil
	case uint32:
		return int64(value), true, nil
	case uint64:
		if value > math.MaxInt64 {
			return invalid()
		}
		return int64(value), true, nil
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) || math.Trunc(value) != value || value < math.MinInt64 || value > math.MaxInt64 {
			return invalid()
		}
		return int64(value), true, nil
	case float32:
		converted := float64(value)
		if math.IsNaN(converted) || math.IsInf(converted, 0) || math.Trunc(converted) != converted || converted < math.MinInt64 || converted > math.MaxInt64 {
			return invalid()
		}
		return int64(converted), true, nil
	case json.Number:
		parsed, err := strconv.ParseInt(strings.TrimSpace(value.String()), 10, 64)
		if err != nil {
			return invalid()
		}
		return parsed, true, nil
	default:
		return invalid()
	}
}
