package tools

import (
	"relay/internal/registry"
)

var defaultRegistry = registry.New()

func DefaultRegistry() *registry.Registry {
	return defaultRegistry
}

func init() {
	registerFileTools()
	registerImageTools()
	registerPDFTools()
	registerDataTools()
	registerTextTools()
	registerWebTools()
	registerWorkflowTools()
}
