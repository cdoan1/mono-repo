package validation

import (
	"fmt"
	"strings"

	"github.com/cdoan1/mono-repo/sdk/featuregate"
	"github.com/cdoan1/mono-repo/sdk/registry"
)

// Operation represents the type of API operation
type Operation string

const (
	OperationCreate Operation = "create"
	OperationUpdate Operation = "update"
)

// Request represents an API request to validate
type Request struct {
	Operation      Operation
	Fields         map[string]interface{}
	FeatureSet     featuregate.FeatureSet
	ExistingFields map[string]interface{}
	EnabledGates   []string
}

// IsFeatureGateEnabled returns true if the given feature gate is enabled for this request
func (r *Request) IsFeatureGateEnabled(gateName string) bool {
	for _, gate := range r.EnabledGates {
		if gate == gateName {
			return true
		}
	}
	return false
}

// ValidationError represents a validation failure
type ValidationError struct {
	FieldPath string
	Reason    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("field %s: %s", e.FieldPath, e.Reason)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []*ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	var sb strings.Builder
	sb.WriteString("validation failed:\n")
	for _, err := range e {
		sb.WriteString("  ")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

// Validator validates API requests against field metadata
type Validator struct {
	registry map[string]registry.FieldMeta
}

// NewValidator creates a validator using the generated field registry
func NewValidator() *Validator {
	return &Validator{
		registry: registry.FieldRegistry,
	}
}

// Validate checks a request against field metadata rules
func (v *Validator) Validate(req *Request) error {
	var errors ValidationErrors

	for fieldPath := range req.Fields {
		meta, exists := v.registry[fieldPath]
		if !exists {
			continue
		}

		if meta.FeatureGate != "" {
			if !featuregate.IsGateEnabled(meta.FeatureGate, req.FeatureSet) {
				errors = append(errors, &ValidationError{
					FieldPath: fieldPath,
					Reason:    fmt.Sprintf("requires feature gate %s which is not enabled in %s feature set", meta.FeatureGate, req.FeatureSet),
				})
				continue
			}
		}

		if err := v.validateWriteMode(fieldPath, meta, req); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func (v *Validator) validateWriteMode(fieldPath string, meta registry.FieldMeta, req *Request) *ValidationError {
	effectiveMode := meta.WriteMode

	if len(meta.FeatureGateAwareWriteModes) > 0 {
		for _, override := range meta.FeatureGateAwareWriteModes {
			if override.FeatureGate != "" && req.IsFeatureGateEnabled(override.FeatureGate) {
				effectiveMode = override.WriteMode
				break
			}
		}

		if effectiveMode == meta.WriteMode {
			for _, override := range meta.FeatureGateAwareWriteModes {
				if override.FeatureGate == "" {
					effectiveMode = override.WriteMode
					break
				}
			}
		}
	}

	switch effectiveMode {
	case registry.ServiceSet:
		return &ValidationError{
			FieldPath: fieldPath,
			Reason:    "field is platform-managed (service-set) and cannot be set by customers",
		}

	case registry.Immutable:
		if req.Operation == OperationUpdate {
			if req.ExistingFields != nil {
				_, existsInOld := req.ExistingFields[fieldPath]
				if existsInOld {
					return &ValidationError{
						FieldPath: fieldPath,
						Reason:    "field is immutable and cannot be changed after creation",
					}
				}
			}
		}
		return nil

	case registry.Mutable:
		return nil

	default:
		return nil
	}
}

// ValidateFieldAccess checks if a customer can access a specific field
func (v *Validator) ValidateFieldAccess(fieldPath string, featureSet featuregate.FeatureSet) error {
	meta, exists := v.registry[fieldPath]
	if !exists {
		return nil
	}

	if meta.FeatureGate != "" {
		if !featuregate.IsGateEnabled(meta.FeatureGate, featureSet) {
			return fmt.Errorf("field %s requires feature gate %s which is not enabled in %s feature set",
				fieldPath, meta.FeatureGate, featureSet)
		}
	}

	return nil
}

// GetFieldMetadata returns metadata for a field path
func (v *Validator) GetFieldMetadata(fieldPath string) (registry.FieldMeta, bool) {
	meta, exists := v.registry[fieldPath]
	return meta, exists
}
