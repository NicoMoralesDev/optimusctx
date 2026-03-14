package extract

import (
	"fmt"
	"sort"

	"github.com/niccrow/optimusctx/internal/repository"
)

type Registry struct {
	adapters map[string]Adapter
}

func NewRegistry(adapters ...Adapter) (Registry, error) {
	registry := Registry{
		adapters: make(map[string]Adapter, len(adapters)),
	}

	for _, adapter := range adapters {
		if adapter == nil {
			return Registry{}, fmt.Errorf("register adapter: adapter is nil")
		}
		language := adapter.Language()
		if language == "" {
			return Registry{}, fmt.Errorf("register adapter: language is required")
		}
		if _, exists := registry.adapters[language]; exists {
			return Registry{}, fmt.Errorf("register adapter: duplicate language %q", language)
		}
		registry.adapters[language] = adapter
	}

	return registry, nil
}

func (r Registry) Resolve(candidate repository.ExtractionCandidate) (Adapter, bool) {
	if candidate.Language == "" {
		return nil, false
	}
	adapter, ok := r.adapters[candidate.Language]
	return adapter, ok
}

func (r Registry) SupportedLanguages() []string {
	languages := make([]string, 0, len(r.adapters))
	for language := range r.adapters {
		languages = append(languages, language)
	}
	sort.Strings(languages)
	return languages
}
