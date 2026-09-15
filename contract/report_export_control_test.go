package contract

import (
	"testing"

	reportmodel "github.com/domainry/domainry-report-sdk/model"
)

func TestReportExportDownloadTTLSecondsIsBusinessAuthoredWithoutDayLimit(t *testing.T) {
	for name, test := range map[string]struct {
		control reportmodel.ReportExportControlSchema
		want    int64
	}{
		"default":             {control: reportmodel.ReportExportControlSchema{}, want: 900},
		"typed seven days":    {control: reportmodel.ReportExportControlSchema{DownloadTTLSeconds: 7 * 24 * 60 * 60}, want: 604800},
		"typed fifteen days":  {control: reportmodel.ReportExportControlSchema{DownloadTTLSeconds: 15 * 24 * 60 * 60}, want: 1296000},
		"legacy fifteen days": {control: reportmodel.ReportExportControlSchema{Config: map[string]any{"download_ttl_seconds": float64(1296000)}}, want: 1296000},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := ReportExportDownloadTTLSeconds(test.control, 900)
			if err != nil || got != test.want {
				t.Fatalf("ttl=%d err=%v want=%d", got, err, test.want)
			}
		})
	}
}

func TestReportExportDownloadTTLSecondsRejectsInvalidOrConflictingValues(t *testing.T) {
	controls := []reportmodel.ReportExportControlSchema{
		{DownloadTTLSeconds: 59},
		{Config: map[string]any{"download_ttl_seconds": 60.5}},
		{Config: map[string]any{"download_ttl_seconds": "604800"}},
		{DownloadTTLSeconds: 604800, Config: map[string]any{"download_ttl_seconds": 1296000}},
	}
	for _, control := range controls {
		if _, err := ReportExportDownloadTTLSeconds(control, 900); err == nil {
			t.Fatalf("invalid control accepted: %#v", control)
		}
	}
}
