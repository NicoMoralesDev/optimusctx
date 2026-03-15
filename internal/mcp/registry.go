package mcp

import (
	"sort"
)

const (
	toolRefresh   = "optimusctx.refresh"
	toolTokenTree = "optimusctx.token_tree"
	toolPack      = "optimusctx.pack"
	toolHealth    = "optimusctx.health"
)

type toolRegistryServices struct {
	query queryToolServices
	ops   operationalToolServices
}

func defaultToolRegistryServices() toolRegistryServices {
	return toolRegistryServices{
		query: defaultQueryToolServices(),
		ops:   defaultOperationalToolServices(),
	}
}

func registerDefaultTools(server *Server) {
	registerTools(server, defaultToolRegistryServices())
}

func registerTools(server *Server, services toolRegistryServices) {
	for _, handler := range toolRegistryHandlers(services) {
		server.RegisterTool(handler)
	}
}

func toolRegistryHandlers(services toolRegistryServices) []ToolHandler {
	handlers := append(queryToolHandlers(services.query), operationalToolHandlers(services.ops)...)
	sort.SliceStable(handlers, func(i, j int) bool {
		return handlers[i].Tool.Name < handlers[j].Tool.Name
	})
	return handlers
}

func toolRegistryNames(services toolRegistryServices) []string {
	handlers := toolRegistryHandlers(services)
	names := make([]string, 0, len(handlers))
	for _, handler := range handlers {
		names = append(names, handler.Tool.Name)
	}
	return names
}
