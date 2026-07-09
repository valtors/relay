package registry

import (
	"context"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_EmptyRegistry(t *testing.T) {
	reg := New()
	assert.Equal(t, 0, reg.Count())
	assert.Empty(t, reg.List())
	assert.Empty(t, reg.ListByCategory("anything"))
}

func TestRegister_SingleTool(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}
	reg.Register("test", mcp.NewTool("single_tool"), handler)

	assert.Equal(t, 1, reg.Count())
	tools := reg.List()
	require.Len(t, tools, 1)
	assert.Equal(t, "single_tool", tools[0].Definition.Name)
	assert.Equal(t, "test", tools[0].Category)
	assert.NotNil(t, tools[0].Handler)
}

func TestRegister_MultipleCategories(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}

	reg.Register("cat_a", mcp.NewTool("tool_a1"), handler)
	reg.Register("cat_a", mcp.NewTool("tool_a2"), handler)
	reg.Register("cat_b", mcp.NewTool("tool_b1"), handler)
	reg.Register("cat_c", mcp.NewTool("tool_c1"), handler)

	assert.Equal(t, 4, reg.Count())

	catA := reg.ListByCategory("cat_a")
	require.Len(t, catA, 2)
	assert.Equal(t, "tool_a1", catA[0].Definition.Name)
	assert.Equal(t, "tool_a2", catA[1].Definition.Name)

	catB := reg.ListByCategory("cat_b")
	require.Len(t, catB, 1)
	assert.Equal(t, "tool_b1", catB[0].Definition.Name)

	catC := reg.ListByCategory("cat_c")
	require.Len(t, catC, 1)
	assert.Equal(t, "tool_c1", catC[0].Definition.Name)
}

func TestListByCategory_NonExistent(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}
	reg.Register("real", mcp.NewTool("tool"), handler)

	result := reg.ListByCategory("nonexistent")
	assert.Empty(t, result)
}

func TestListByCategory_EmptyCategoryString(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}
	reg.Register("", mcp.NewTool("empty_cat_tool"), handler)
	reg.Register("real", mcp.NewTool("real_cat_tool"), handler)

	emptyCat := reg.ListByCategory("")
	require.Len(t, emptyCat, 1)
	assert.Equal(t, "empty_cat_tool", emptyCat[0].Definition.Name)
}

func TestList_ReturnsCopy(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}
	reg.Register("test", mcp.NewTool("tool1"), handler)

	first := reg.List()
	second := reg.List()
	assert.Equal(t, len(first), len(second))

	reg.Register("test", mcp.NewTool("tool2"), handler)
	assert.Equal(t, 1, len(first), "original slice should not be affected by new registration")
	assert.Equal(t, 2, reg.Count())
}

func TestList_PreservesOrder(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}

	names := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	for _, n := range names {
		reg.Register("ordered", mcp.NewTool(n), handler)
	}

	tools := reg.List()
	require.Len(t, tools, len(names))
	for i, expected := range names {
		assert.Equal(t, expected, tools[i].Definition.Name, "order should be preserved at index %d", i)
	}
}

func TestRegister_Concurrent(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}

	var wg sync.WaitGroup
	goroutines := 50
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			reg.Register("concurrent", mcp.NewTool("tool"), handler)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, goroutines, reg.Count())
}

func TestList_ConcurrentReads(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}
	for i := 0; i < 20; i++ {
		reg.Register("test", mcp.NewTool("tool"), handler)
	}

	var wg sync.WaitGroup
	goroutines := 50
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			tools := reg.List()
			assert.Len(t, tools, 20)
			byCat := reg.ListByCategory("test")
			assert.Len(t, byCat, 20)
		}()
	}
	wg.Wait()
}

func TestCount_AfterMultipleRegistrations(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}

	assert.Equal(t, 0, reg.Count())
	reg.Register("a", mcp.NewTool("t1"), handler)
	assert.Equal(t, 1, reg.Count())
	reg.Register("b", mcp.NewTool("t2"), handler)
	assert.Equal(t, 2, reg.Count())
	reg.Register("c", mcp.NewTool("t3"), handler)
	assert.Equal(t, 3, reg.Count())
}

func TestListByCategory_AllToolsSameCategory(t *testing.T) {
	reg := New()
	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}

	reg.Register("only_cat", mcp.NewTool("t1"), handler)
	reg.Register("only_cat", mcp.NewTool("t2"), handler)
	reg.Register("only_cat", mcp.NewTool("t3"), handler)

	all := reg.ListByCategory("only_cat")
	assert.Len(t, all, 3)
	other := reg.ListByCategory("other_cat")
	assert.Empty(t, other)
}
