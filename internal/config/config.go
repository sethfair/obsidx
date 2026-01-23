package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// TagWeight defines a weight multiplier for a specific tag
type TagWeight struct {
	Tag    string  `json:"tag"`
	Weight float32 `json:"weight"`
}

// StatusWeight defines a weight multiplier for a status value
type StatusWeight struct {
	Status string  `json:"status"`
	Weight float32 `json:"weight"`
}

// WeightConfig holds the customizable weighting rules
type WeightConfig struct {
	// Tag-based weights (e.g., #permanent-note, #fleeting-notes)
	TagWeights []TagWeight `json:"tag_weights"`

	// Status-based weights
	StatusWeights []StatusWeight `json:"status_weights"`

	// Default weight when no tags match
	DefaultWeight float32 `json:"default_weight"`

	// Whether to multiply all matching tag weights (true) or use max (false)
	MultiplyTagWeights bool `json:"multiply_tag_weights"`
}

// DefaultWeightConfig returns sensible defaults
func DefaultWeightConfig() *WeightConfig {
	return &WeightConfig{
		TagWeights: []TagWeight{
			// Zettelkasten system - permanent notes are most valuable
			{Tag: "permanent-note", Weight: 1.3},
			{Tag: "literature-note", Weight: 1.1},
			{Tag: "fleeting-notes", Weight: 0.8},
			{Tag: "reference", Weight: 1.0},

			// PARA - Projects are high priority, archives low
			{Tag: "writerflow", Weight: 1.2},
			{Tag: "mvp-1", Weight: 1.2},
			{Tag: "product-development", Weight: 1.15},
			{Tag: "marketing", Weight: 1.1},
			{Tag: "product", Weight: 1.1},
			{Tag: "business", Weight: 1.1},
			{Tag: "archive", Weight: 0.6},

			// Writerflow - high-value customer/product insights
			{Tag: "customer-research", Weight: 1.25},
			{Tag: "validation", Weight: 1.2},
			{Tag: "mom-test", Weight: 1.2},
			{Tag: "icp", Weight: 1.25},
			{Tag: "vision", Weight: 1.3},
			{Tag: "positioning", Weight: 1.15},
		},
		StatusWeights: []StatusWeight{
			{Status: "active", Weight: 1.0},
			{Status: "draft", Weight: 0.9},
			{Status: "superseded", Weight: 0.5},
			{Status: "deprecated", Weight: 0.5},
		},
		DefaultWeight:      1.0,
		MultiplyTagWeights: false, // use max weight by default
	}
}

// LoadWeightConfig loads configuration from a JSON file
// Falls back to defaults if file doesn't exist
func LoadWeightConfig(configPath string) (*WeightConfig, error) {
	// If no config file, use defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultWeightConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config WeightConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults for missing fields
	if config.DefaultWeight == 0 {
		config.DefaultWeight = 1.0
	}

	return &config, nil
}

// SaveWeightConfig saves configuration to a JSON file
func SaveWeightConfig(config *WeightConfig, configPath string) error {
	// Create directory if needed
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// CalculateWeight computes the final weight for a note based on tags and status
func (c *WeightConfig) CalculateWeight(tags []string, status string) float32 {
	// Calculate tag weight
	tagWeight := c.DefaultWeight
	matchedAny := false

	if c.MultiplyTagWeights {
		// Multiply all matching tag weights
		tagWeight = 1.0
		for _, tag := range tags {
			for _, tw := range c.TagWeights {
				if matchTag(tag, tw.Tag) {
					tagWeight *= tw.Weight
					matchedAny = true
				}
			}
		}
		if !matchedAny {
			tagWeight = c.DefaultWeight
		}
	} else {
		// Use maximum weight among matching tags
		for _, tag := range tags {
			for _, tw := range c.TagWeights {
				if matchTag(tag, tw.Tag) {
					if tw.Weight > tagWeight {
						tagWeight = tw.Weight
					}
					matchedAny = true
				}
			}
		}
	}

	// Calculate status weight
	statusWeight := 1.0
	for _, sw := range c.StatusWeights {
		if sw.Status == status {
			statusWeight = float64(sw.Weight)
			break
		}
	}

	return tagWeight * float32(statusWeight)
}

// matchTag checks if a tag matches, handling # prefix variations
func matchTag(noteTag, configTag string) bool {
	// Normalize both to remove # prefix
	noteTag = normalizeTag(noteTag)
	configTag = normalizeTag(configTag)
	return noteTag == configTag
}

func normalizeTag(tag string) string {
	if len(tag) > 0 && tag[0] == '#' {
		return tag[1:]
	}
	return tag
}
