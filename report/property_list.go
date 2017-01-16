package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/common/mtime"
)

// MaxPropertyListSize sets the limit on the size of the property list to render.
// TODO: this won't be needed once we send reports incrementally
const (
	MaxPropertyListSize   = 20
	TruncationCountPrefix = "property_list_truncation_count_"
)

// AddPrefixPropertyList appends arbitrary key-value pairs to the Node, returning a new node.
func (node Node) AddPrefixPropertyList(prefix string, properties map[string]string) Node {
	count := 0
	for key, value := range properties {
		if count >= MaxPropertyListSize {
			break
		}
		node = node.WithLatest(prefix+key, mtime.Now(), value)
		count++
	}
	if len(properties) > MaxPropertyListSize {
		truncationCount := fmt.Sprintf("%d", len(properties)-MaxPropertyListSize)
		node = node.WithLatest(TruncationCountPrefix+prefix, mtime.Now(), truncationCount)
	}
	return node
}

// ExtractPropertyList returns the key-value pairs to build a property list from this node
func (node Node) ExtractPropertyList(template PropertyListTemplate) (rows map[string]string, truncationCount int) {
	rows = map[string]string{}
	truncationCount = 0
	node.Latest.ForEach(func(key string, _ time.Time, value string) {
		if label, ok := template.FixedProperties[key]; ok {
			rows[label] = value
		}
		if len(template.Prefix) > 0 && strings.HasPrefix(key, template.Prefix) {
			label := key[len(template.Prefix):]
			rows[label] = value
		}
	})
	if str, ok := node.Latest.Lookup(TruncationCountPrefix + template.Prefix); ok {
		if n, err := fmt.Sscanf(str, "%d", &truncationCount); n != 1 || err != nil {
			log.Warn("Unexpected truncation count format %q", str)
		}
	}
	return rows, truncationCount
}

// PropertyList is the type for a property list (labels) in the UI.
type PropertyList struct {
	ID              string        `json:"id"`
	Label           string        `json:"label"`
	Rows            []MetadataRow `json:"rows"`
	TruncationCount int           `json:"truncationCount,omitempty"`
}

type propertyListsByID []PropertyList

func (t propertyListsByID) Len() int           { return len(t) }
func (t propertyListsByID) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t propertyListsByID) Less(i, j int) bool { return t[i].ID < t[j].ID }

// Copy returns a copy of the PropertyList.
func (t PropertyList) Copy() PropertyList {
	result := PropertyList{
		ID:    t.ID,
		Label: t.Label,
		Rows:  make([]MetadataRow, 0, len(t.Rows)),
	}
	for _, row := range t.Rows {
		result.Rows = append(result.Rows, row)
	}
	return result
}

// PropertyListTemplate describes how to render a property list for the UI.
type PropertyListTemplate struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Prefix string `json:"prefix"`
	// FixedProperties indicates what predetermined properties to render each entry is
	// indexed by the key to extract the row value is mapped to the row
	// label
	FixedProperties map[string]string `json:"fixedProperties"`
}

// Copy returns a value-copy of the PropertyListTemplate
func (t PropertyListTemplate) Copy() PropertyListTemplate {
	FixedPropertiesCopy := make(map[string]string, len(t.FixedProperties))
	for key, value := range t.FixedProperties {
		FixedPropertiesCopy[key] = value
	}
	t.FixedProperties = FixedPropertiesCopy
	return t
}

// Merge other into t, returning a fresh copy.  Does fieldwise max -
// whilst this isn't particularly meaningful, at least it idempotent,
// commutativite and associative.
func (t PropertyListTemplate) Merge(other PropertyListTemplate) PropertyListTemplate {
	max := func(s1, s2 string) string {
		if s1 > s2 {
			return s1
		}
		return s2
	}

	fixedProperties := t.FixedProperties
	if len(other.FixedProperties) > len(fixedProperties) {
		fixedProperties = other.FixedProperties
	}

	return PropertyListTemplate{
		ID:              max(t.ID, other.ID),
		Label:           max(t.Label, other.Label),
		Prefix:          max(t.Prefix, other.Prefix),
		FixedProperties: fixedProperties,
	}
}

// PropertyListTemplates is a mergeable set of PropertyListTemplate
type PropertyListTemplates map[string]PropertyListTemplate

// PropertyLists renders the PropertyListTemplates for a given node.
func (t PropertyListTemplates) PropertyLists(node Node) []PropertyList {
	var result []PropertyList
	for _, template := range t {
		rows, truncationCount := node.ExtractPropertyList(template)
		propertyList := PropertyList{
			ID:              template.ID,
			Label:           template.Label,
			Rows:            []MetadataRow{},
			TruncationCount: truncationCount,
		}
		keys := make([]string, 0, len(rows))
		for k := range rows {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			propertyList.Rows = append(propertyList.Rows, MetadataRow{
				ID:    "label_" + key,
				Label: key,
				Value: rows[key],
			})
		}
		result = append(result, propertyList)
	}
	sort.Sort(propertyListsByID(result))
	return result
}

// Copy returns a value copy of the PropertyListTemplates
func (t PropertyListTemplates) Copy() PropertyListTemplates {
	if t == nil {
		return nil
	}
	result := PropertyListTemplates{}
	for k, v := range t {
		result[k] = v.Copy()
	}
	return result
}

// Merge merges two sets of PropertyListTemplates
func (t PropertyListTemplates) Merge(other PropertyListTemplates) PropertyListTemplates {
	if t == nil && other == nil {
		return nil
	}
	result := make(PropertyListTemplates, len(t))
	for k, v := range t {
		result[k] = v
	}
	for k, v := range other {
		if existing, ok := result[k]; ok {
			result[k] = v.Merge(existing)
		} else {
			result[k] = v
		}
	}
	return result
}
