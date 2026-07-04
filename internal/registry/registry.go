package registry

import (
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Registry struct {
	mu    sync.RWMutex
	tools []Tool
}

type Tool struct {
	Definition mcp.Tool
	Handler    server.ToolHandlerFunc
	Category   string
}

func New() *Registry {
	return &Registry{
		tools: make([]Tool, 0),
	}
}

func (r *Registry) Register(category string, def mcp.Tool, handler server.ToolHandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = append(r.tools, Tool{
		Definition: def,
		Handler:    handler,
		Category:   category,
	})
}

func (r *Registry) RegisterAll(s *server.MCPServer) {
	for _, tool := range r.List() {
		s.AddTool(tool.Definition, tool.Handler)
	}
}

func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, len(r.tools))
	copy(tools, r.tools)
	return tools
}

func (r *Registry) ListByCategory(cat string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0)
	for _, tool := range r.tools {
		if tool.Category == cat {
			tools = append(tools, tool)
		}
	}
	return tools
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}
