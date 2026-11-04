package plan

import (
	"sort"
	"strings"
)

// defaultRegistryPrefix is the host and separator of the public Terraform registry.
const defaultRegistryPrefix = "registry.terraform.io/"

// normalizeProviderName reduces a Terraform provider address to its type name,
// e.g. "registry.terraform.io/hashicorp/googlebeta" -> "googlebeta". Legacy
// "provider.aws" and bare "aws" forms are also supported.
func normalizeProviderName(providerName string) string {
	name := strings.TrimSpace(providerName)
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return strings.TrimPrefix(name, "provider.")
}

// providerGroupKey returns the identity used to group a change by provider. It is
// the full provider address when known, so providers that share a type name but
// come from different namespaces or registries stay separate. It falls back to the
// short provider name, and then to the resource type prefix.
func providerGroupKey(change ResourceChange) string {
	if change.ProviderAddress != "" {
		return change.ProviderAddress
	}
	if change.Provider != "" {
		return change.Provider
	}
	if prefix, _, _ := strings.Cut(change.Type, "_"); prefix != "" {
		return prefix
	}
	return "unknown"
}

// providerGroupLabels maps provider group keys to display labels. Labels use the
// short provider name, unless several keys share it; those use the address
// without the default registry host (e.g. "hashicorp/widget" and "acme/widget").
func providerGroupLabels(keys []string) map[string]string {
	shortCounts := make(map[string]int, len(keys))
	for _, key := range keys {
		shortCounts[normalizeProviderName(key)]++
	}

	labels := make(map[string]string, len(keys))
	for _, key := range keys {
		short := normalizeProviderName(key)
		if shortCounts[short] > 1 {
			labels[key] = strings.TrimPrefix(key, defaultRegistryPrefix)
		} else {
			labels[key] = short
		}
	}
	return labels
}

// sortedProviderGroups returns the group keys ordered by display label (then key)
// for deterministic output, together with the display label for each key.
func sortedProviderGroups(groups map[string][]ResourceChange) ([]string, map[string]string) {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	labels := providerGroupLabels(keys)
	sort.Slice(keys, func(i, j int) bool {
		if labels[keys[i]] != labels[keys[j]] {
			return labels[keys[i]] < labels[keys[j]]
		}
		return keys[i] < keys[j]
	})
	return keys, labels
}
