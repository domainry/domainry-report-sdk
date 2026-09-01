// Package contract owns deployment-neutral Report definition contracts shared
// by authoring, embedded hosts, SaaS owners, and asynchronous workers.
package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// CanonicalJSONSHA256 returns the lowercase SHA-256 digest of Go's stable JSON
// encoding for a Report contract value.
func CanonicalJSONSHA256(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return SHA256Hex(encoded), nil
}

func SHA256Hex(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

// SafeExportFilename derives the portable CSV filename used by Report and
// Data Exchange without exposing any host storage naming.
func SafeExportFilename(reportKey, objectKey string) string {
	clean := func(value string) string {
		value = strings.TrimSpace(value)
		var result strings.Builder
		for _, char := range value {
			if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '_' {
				result.WriteRune(char)
			} else {
				result.WriteByte('_')
			}
		}
		return strings.Trim(result.String(), "_")
	}
	return clean(reportKey) + "-" + clean(objectKey) + ".csv"
}
