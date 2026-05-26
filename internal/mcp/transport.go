package mcp

import "github.com/mark3labs/mcp-go/server"

func StartStdio(s *server.MCPServer) error {
	return server.ServeStdio(s)
}
