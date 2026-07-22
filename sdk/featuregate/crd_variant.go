package featuregate

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/cdoan1/mono-repo/sdk/registry"
)

// CRDVariantGenerator generates feature-set-specific CRD variants
type CRDVariantGenerator struct {
	fieldRegistry map[string]registry.FieldMeta
}

// NewCRDVariantGenerator creates a new CRD variant generator
func NewCRDVariantGenerator() *CRDVariantGenerator {
	return &CRDVariantGenerator{
		fieldRegistry: registry.FieldRegistry,
	}
}

// GenerateVariant reads a base CRD and generates a filtered variant for a feature set
func (g *CRDVariantGenerator) GenerateVariant(inputPath string, outputPath string, featureSet FeatureSet) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading CRD: %w", err)
	}

	var crd yaml.Node
	if err := yaml.Unmarshal(data, &crd); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	ctx := &filterContext{
		featureSet: featureSet,
		inSchema:   false,
		fieldPath:  "",
	}
	if err := g.filterCRDNode(&crd, ctx); err != nil {
		return fmt.Errorf("filtering CRD: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer func() { _ = f.Close() }()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(&crd); err != nil {
		return fmt.Errorf("writing YAML: %w", err)
	}

	return nil
}

type filterContext struct {
	featureSet FeatureSet
	inSchema   bool
	fieldPath  string
}

func (g *CRDVariantGenerator) filterCRDNode(node *yaml.Node, ctx *filterContext) error {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if err := g.filterCRDNode(child, ctx); err != nil {
				return err
			}
		}

	case yaml.MappingNode:
		newContent := make([]*yaml.Node, 0, len(node.Content))

		for i := 0; i < len(node.Content); i += 2 {
			if i+1 >= len(node.Content) {
				break
			}

			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			fieldName := keyNode.Value

			enteringSchema := !ctx.inSchema && fieldName == "properties"

			oldInSchema := ctx.inSchema
			oldFieldPath := ctx.fieldPath

			if enteringSchema {
				ctx.inSchema = true
			} else if ctx.inSchema && fieldName != "properties" {
				if ctx.fieldPath == "" {
					ctx.fieldPath = fieldName
				} else {
					ctx.fieldPath = ctx.fieldPath + "." + fieldName
				}
			}

			shouldInclude := true
			if ctx.inSchema && fieldName != "properties" && ctx.fieldPath != "" {
				shouldInclude = g.shouldIncludeField(ctx.fieldPath, ctx.featureSet)
			}

			if shouldInclude {
				if err := g.filterCRDNode(valueNode, ctx); err != nil {
					return err
				}
				newContent = append(newContent, keyNode, valueNode)
			}

			ctx.inSchema = oldInSchema
			ctx.fieldPath = oldFieldPath
		}

		node.Content = newContent

	case yaml.SequenceNode:
		for _, child := range node.Content {
			if err := g.filterCRDNode(child, ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *CRDVariantGenerator) shouldIncludeField(fieldPath string, featureSet FeatureSet) bool {
	meta, exists := g.fieldRegistry[fieldPath]
	if !exists {
		return true
	}

	if meta.FeatureGate != "" {
		return IsGateEnabled(meta.FeatureGate, featureSet)
	}

	return true
}

// GenerateAllVariants generates CRD variants for all feature sets
func (g *CRDVariantGenerator) GenerateAllVariants(inputPath string, outputDir string, baseName string) error {
	featureSets := []struct {
		set    FeatureSet
		suffix string
	}{
		{Default, "default"},
		{TechPreviewNoUpgrade, "techpreview"},
		{DevPreviewNoUpgrade, "devpreview"},
	}

	for _, fs := range featureSets {
		outputPath := fmt.Sprintf("%s/%s_%s.yaml", outputDir, baseName, fs.suffix)
		if err := g.GenerateVariant(inputPath, outputPath, fs.set); err != nil {
			return fmt.Errorf("generating %s variant: %w", fs.suffix, err)
		}
	}

	return nil
}

// WriteVariantToWriter generates a variant and writes it to a writer
func (g *CRDVariantGenerator) WriteVariantToWriter(inputPath string, w io.Writer, featureSet FeatureSet) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading CRD: %w", err)
	}

	var crd yaml.Node
	if err := yaml.Unmarshal(data, &crd); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	ctx := &filterContext{
		featureSet: featureSet,
		inSchema:   false,
		fieldPath:  "",
	}
	if err := g.filterCRDNode(&crd, ctx); err != nil {
		return fmt.Errorf("filtering CRD: %w", err)
	}

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	if err := encoder.Encode(&crd); err != nil {
		return fmt.Errorf("writing YAML: %w", err)
	}

	return nil
}
