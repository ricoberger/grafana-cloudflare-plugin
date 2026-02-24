package cloudflare

import (
	"regexp"
	"strings"
	"time"
)

var legendReplacer = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)

type Volume struct {
	Name      string
	Filter    string
	Dimension string
}

type FrameData struct {
	Index      int
	Name       string
	Timestamps []time.Time
	Values     []float64
	Labels     map[string]string
}

func parseLegend(name, legend string, labels map[string]string) string {
	if legend == "" {
		return name
	}

	result := legendReplacer.ReplaceAllStringFunc(legend, func(in string) string {
		labelName := strings.Replace(in, "{{", "", 1)
		labelName = strings.Replace(labelName, "}}", "", 1)
		labelName = strings.TrimSpace(labelName)
		if val, ok := labels[labelName]; ok {
			return val
		}
		return ""
	})
	if result == "" {
		return name
	}
	return result
}
