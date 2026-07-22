package featuregate

import (
	"fmt"

	"github.com/cdoan1/mono-repo/sdk/registry"
)

// FilterCRDFields returns a list of field paths that should be included in the CRD
// for the given feature set
func FilterCRDFields(featureSet FeatureSet) []string {
	var includedFields []string

	for fieldPath, meta := range registry.FieldRegistry {
		if meta.Hidden {
			continue
		}

		if meta.FeatureGate == "" {
			includedFields = append(includedFields, fieldPath)
			continue
		}

		if IsGateEnabled(meta.FeatureGate, featureSet) {
			includedFields = append(includedFields, fieldPath)
		}
	}

	return includedFields
}

// FieldsForFeatureSet returns field metadata for all fields available in the given feature set
func FieldsForFeatureSet(featureSet FeatureSet) map[string]registry.FieldMeta {
	result := make(map[string]registry.FieldMeta)

	for fieldPath, meta := range registry.FieldRegistry {
		if meta.Hidden {
			continue
		}

		if meta.FeatureGate == "" {
			result[fieldPath] = meta
			continue
		}

		if IsGateEnabled(meta.FeatureGate, featureSet) {
			result[fieldPath] = meta
		}
	}

	return result
}

// SummarizeFeatureSet returns a summary of fields available in each feature set
func SummarizeFeatureSet() string {
	featureSets := []FeatureSet{Default, TechPreviewNoUpgrade, DevPreviewNoUpgrade}

	summary := "Feature Set Field Summary:\n\n"

	for _, fs := range featureSets {
		fields := FieldsForFeatureSet(fs)
		gates := GatesForFeatureSet(fs)

		summary += fmt.Sprintf("%s:\n", fs)
		summary += fmt.Sprintf("  Total fields: %d\n", len(fields))
		summary += fmt.Sprintf("  Enabled gates: %v\n", gates)
		summary += "\n"
	}

	return summary
}
