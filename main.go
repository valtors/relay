package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"relay/internal/license"
	"relay/internal/logger"
	"relay/tools"
)

var Version = "dev"

func main() {
	httpMode := flag.Bool("http", false,
		"serve over Streamable-HTTP (Claude.ai compatible) instead of stdio")
	addr := flag.String("addr", ":8080",
		"HTTP listen address (only used with -http). Endpoint is <addr>/mcp")
	flag.Parse()

	_ = godotenv.Load()

	lic, err := license.Verify()
	if err != nil {
		fmt.Fprint(os.Stderr, license.FriendlyMessage(err))
		os.Exit(1)
	}
	logger.Info("licensed", "subject", lic.Subject, "expires", lic.Expires)

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Fprintln(os.Stderr,
			"\n[relay] ANTHROPIC_API_KEY is not set."+
				"\nAdd it to your shell environment or a .env file in the project root.",
		)
		os.Exit(1)
	}

	s := buildServer()

	if *httpMode {
		serveHTTP(s, *addr)
		return
	}

	logger.Info("relay starting", "version", Version, "transport", "stdio")
	if err := server.ServeStdio(s); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}

func buildServer() *server.MCPServer {
	s := server.NewMCPServer(
		"relay",
		Version,
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("run_workflow",
		mcp.WithDescription(
			"Run the full multi-agent product launch pipeline end-to-end. "+
				"Reads product_brief.md from the working directory. "+
				"Orchestrates PM Plan → Research → Brand → UX → GTM with human review after each stage. "+
				"Auto-resumes from the last completed stage if called again after a crash. "+
				"All outputs written to ./output/.",
		),
		mcp.WithString("brief_path",
			mcp.Description("Path to brief file. Defaults to ./product_brief.md"),
		),
		mcp.WithBoolean("resume",
			mcp.Description("Force resume from last stage. Default: true if session exists"),
		),
	), tools.RunWorkflow)

	s.AddTool(mcp.NewTool("pm_plan",
		mcp.WithDescription(
			"PM Agent reads the product brief and writes a focused brief for Agent 1 (Research). "+
				"Run this first.",
		),
		mcp.WithString("brief_path",
			mcp.Description("Path to brief file. Defaults to ./product_brief.md"),
		),
	), tools.PMPlan)

	s.AddTool(mcp.NewTool("run_research",
		mcp.WithDescription(
			"Agent 1: Market research, ICP classification, competitor analysis. "+
				"Uses web search. Reads pm_brief_for_agent1.md from ./output/. "+
				"Writes 01_research.md.",
		),
		mcp.WithString("extra_notes",
			mcp.Description("Notes from a prior iterate decision to incorporate"),
		),
	), tools.RunResearch)

	s.AddTool(mcp.NewTool("run_brand",
		mcp.WithDescription(
			"Agent 2: Positioning statement, brand voice, messaging pillars. "+
				"Reads 01_research.md. Writes 02_brand_messaging.md.",
		),
		mcp.WithString("extra_notes", mcp.Description("Iteration notes")),
	), tools.RunBrand)

	s.AddTool(mcp.NewTool("run_ux",
		mcp.WithDescription(
			"Agent 3: Wireframe briefs, screen list, user flows, image-prototype prompts. "+
				"Writes 03_ux.md.",
		),
		mcp.WithString("extra_notes", mcp.Description("Iteration notes")),
	), tools.RunUX)

	s.AddTool(mcp.NewTool("run_gtm",
		mcp.WithDescription(
			"Agent 4: Social media (4a) and B2B outreach (4b) run in parallel via goroutines. "+
				"Writes 04_go_to_market.md.",
		),
		mcp.WithString("extra_notes", mcp.Description("Iteration notes")),
	), tools.RunGTM)

	s.AddTool(mcp.NewTool("request_approval",
		mcp.WithDescription(
			"Present a stage summary to the human and wait for their approve or iterate decision. "+
				"Blocks until input is received on stdin.",
		),
		mcp.WithString("checkpoint",
			mcp.Required(),
			mcp.Description("H1, H2, H3, or H4"),
		),
		mcp.WithString("summary",
			mcp.Required(),
			mcp.Description("PM Agent summary — 5 to 10 bullets"),
		),
		mcp.WithArray("questions",
			mcp.Required(),
			mcp.Description("2-3 specific questions for the human"),
			mcp.Items(map[string]any{"type": "string"}),
		),
	), tools.RequestApproval)

	s.AddTool(mcp.NewTool("assemble_plan",
		mcp.WithDescription(
			"PM Agent assembles all stage outputs into final_product_plan.md. "+
				"Call after H4 is approved.",
		),
		mcp.WithString("product_name", mcp.Description("Optional product name for the title")),
	), tools.AssemblePlan)

	return s
}

func serveHTTP(s *server.MCPServer, addr string) {
	httpSrv := server.NewStreamableHTTPServer(s)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("relay starting",
			"version", Version, "transport", "streamable-http",
			"addr", addr, "endpoint", addr+"/mcp",
		)
		if err := httpSrv.Start(addr); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("http server error", "err", err)
			os.Exit(1)
		}
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "err", err)
			os.Exit(1)
		}
	}
}
