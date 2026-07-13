package registry

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterListByCategoryCountAndRegisterAll(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}

	reg.Register("workflow", mcp.NewTool("tool_one"), handler)
	reg.Register("workflow", mcp.NewTool("tool_two"), handler)
	reg.Register("text", mcp.NewTool("tool_three"), handler)

	assert.Equal(t, 3, reg.Count())

	all := reg.List()
	require.Len(t, all, 3)

	workflowTools := reg.ListByCategory("workflow")
	require.Len(t, workflowTools, 2)
	assert.Equal(t, "tool_one", workflowTools[0].Definition.Name)
	assert.Equal(t, "tool_two", workflowTools[1].Definition.Name)

	textTools := reg.ListByCategory("text")
	require.Len(t, textTools, 1)
	assert.Equal(t, "tool_three", textTools[0].Definition.Name)

	s := server.NewMCPServer("test", "dev", server.WithToolCapabilities(true))
	reg.RegisterAll(s)

	c, err := client.NewInProcessClient(s)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = c.Initialize(ctx, mcp.InitializeRequest{})
	require.NoError(t, err)

	resp, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Tools, 3)

	names := []string{resp.Tools[0].Name, resp.Tools[1].Name, resp.Tools[2].Name}
	assert.ElementsMatch(t, []string{"tool_one", "tool_two", "tool_three"}, names)
}

func TestRegistry_RegisterAllDispatchesHandler(t *testing.T) {
	reg := New()
	reg.Register("workflow", mcp.NewTool("echo"), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		message, _ := args["message"].(string)
		return mcp.NewToolResultText("handled:" + message), nil
	})

	s := server.NewMCPServer("test", "dev", server.WithToolCapabilities(true))
	reg.RegisterAll(s)

	c, err := client.NewInProcessClient(s)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = c.Initialize(ctx, mcp.InitializeRequest{})
	require.NoError(t, err)

	resp, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "echo",
			Arguments: map[string]any{"message": "registry-ok"},
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Content, 1)

	text, ok := resp.Content[0].(mcp.TextContent)
	require.True(t, ok, "expected text content, got %T", resp.Content[0])
	assert.Equal(t, "handled:registry-ok", text.Text)

	_, err = c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "missing"},
	})
	require.Error(t, err)
}

func TestRegistry_RegisterDuplicateNamesAndNilHandlers(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}

	reg.Register("workflow", mcp.NewTool("duplicate"), handler)
	reg.Register("workflow", mcp.NewTool("duplicate"), nil)

	tools := reg.List()
	require.Len(t, tools, 2)
	assert.Equal(t, "duplicate", tools[0].Definition.Name)
	assert.Equal(t, "duplicate", tools[1].Definition.Name)
	assert.NotNil(t, tools[0].Handler)
	assert.Nil(t, tools[1].Handler)

	workflowTools := reg.ListByCategory("workflow")
	require.Len(t, workflowTools, 2)
	assert.NotNil(t, workflowTools[0].Handler)
	assert.Nil(t, workflowTools[1].Handler)
}
