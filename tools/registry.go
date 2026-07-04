package tools

import (
	"github.com/mark3labs/mcp-go/mcp"

	"relay/internal/registry"
)

var defaultRegistry = registry.New()

func DefaultRegistry() *registry.Registry {
	return defaultRegistry
}

func init() {
	defaultRegistry.Register("workflow", mcp.NewTool("run_workflow",
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
	), RunWorkflow)

	defaultRegistry.Register("workflow", mcp.NewTool("pm_plan",
		mcp.WithDescription(
			"PM Agent reads the product brief and writes a focused brief for Agent 1 (Research). "+
				"Run this first.",
		),
		mcp.WithString("brief_path",
			mcp.Description("Path to brief file. Defaults to ./product_brief.md"),
		),
	), PMPlan)

	defaultRegistry.Register("workflow", mcp.NewTool("run_research",
		mcp.WithDescription(
			"Agent 1: Market research, ICP classification, competitor analysis. "+
				"Uses web search. Reads pm_brief_for_agent1.md from ./output/. "+
				"Writes 01_research.md.",
		),
		mcp.WithString("extra_notes",
			mcp.Description("Notes from a prior iterate decision to incorporate"),
		),
	), RunResearch)

	defaultRegistry.Register("workflow", mcp.NewTool("run_brand",
		mcp.WithDescription(
			"Agent 2: Positioning statement, brand voice, messaging pillars. "+
				"Reads 01_research.md. Writes 02_brand_messaging.md.",
		),
		mcp.WithString("extra_notes", mcp.Description("Iteration notes")),
	), RunBrand)

	defaultRegistry.Register("workflow", mcp.NewTool("run_ux",
		mcp.WithDescription(
			"Agent 3: Wireframe briefs, screen list, user flows, image-prototype prompts. "+
				"Writes 03_ux.md.",
		),
		mcp.WithString("extra_notes", mcp.Description("Iteration notes")),
	), RunUX)

	defaultRegistry.Register("workflow", mcp.NewTool("run_gtm",
		mcp.WithDescription(
			"Agent 4: Social media (4a) and B2B outreach (4b) run in parallel via goroutines. "+
				"Writes 04_go_to_market.md.",
		),
		mcp.WithString("extra_notes", mcp.Description("Iteration notes")),
	), RunGTM)

	defaultRegistry.Register("workflow", mcp.NewTool("request_approval",
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
	), RequestApproval)

	defaultRegistry.Register("workflow", mcp.NewTool("assemble_plan",
		mcp.WithDescription(
			"PM Agent assembles all stage outputs into final_product_plan.md. "+
				"Call after H4 is approved.",
		),
		mcp.WithString("product_name", mcp.Description("Optional product name for the title")),
	), AssemblePlan)

	defaultRegistry.Register("file", mcp.NewTool("file_hash",
		mcp.WithDescription("Compute a file SHA-256 hash"),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
	), FileHash)
	defaultRegistry.Register("file", mcp.NewTool("file_read",
		mcp.WithDescription("Read file contents"),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
	), FileRead)
	defaultRegistry.Register("file", mcp.NewTool("file_write",
		mcp.WithDescription("Write content to a file"),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
		mcp.WithString("content", mcp.Required(), mcp.Description("File content")),
	), FileWrite)
	defaultRegistry.Register("file", mcp.NewTool("file_list",
		mcp.WithDescription("List directory contents"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Directory path")),
		mcp.WithBoolean("recursive", mcp.Description("List recursively")),
	), FileList)
	defaultRegistry.Register("file", mcp.NewTool("file_size",
		mcp.WithDescription("Get file size"),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
	), FileSize)
	defaultRegistry.Register("file", mcp.NewTool("file_zip",
		mcp.WithDescription("Create a zip archive"),
		mcp.WithArray("paths",
			mcp.Required(),
			mcp.Description("Paths to archive"),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithString("output", mcp.Required(), mcp.Description("Archive path")),
	), FileZip)
	defaultRegistry.Register("file", mcp.NewTool("file_unzip",
		mcp.WithDescription("Extract a zip archive"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Zip file path")),
		mcp.WithString("output_dir", mcp.Description("Output directory")),
	), FileUnzip)

	defaultRegistry.Register("image", mcp.NewTool("image_info",
		mcp.WithDescription("Get image metadata including dimensions, format, and file size"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Image path")),
	), ImageInfo)
	defaultRegistry.Register("image", mcp.NewTool("image_resize",
		mcp.WithDescription("Resize an image using nearest-neighbor scaling"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Image path")),
		mcp.WithNumber("width", mcp.Description("Target width in pixels")),
		mcp.WithNumber("height", mcp.Description("Target height in pixels")),
		mcp.WithString("output", mcp.Description("Optional output path")),
	), ImageResize)
	defaultRegistry.Register("image", mcp.NewTool("image_crop",
		mcp.WithDescription("Crop an image to the given rectangle"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Image path")),
		mcp.WithNumber("x", mcp.Required(), mcp.Description("Left coordinate in pixels")),
		mcp.WithNumber("y", mcp.Required(), mcp.Description("Top coordinate in pixels")),
		mcp.WithNumber("width", mcp.Required(), mcp.Description("Crop width in pixels")),
		mcp.WithNumber("height", mcp.Required(), mcp.Description("Crop height in pixels")),
		mcp.WithString("output", mcp.Description("Optional output path")),
	), ImageCrop)
	defaultRegistry.Register("image", mcp.NewTool("image_convert",
		mcp.WithDescription("Convert an image to png, jpeg, or gif"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Image path")),
		mcp.WithString("format", mcp.Required(), mcp.Description("Target format: png, jpeg, or gif")),
		mcp.WithNumber("quality", mcp.Description("JPEG quality from 1-100; default 85")),
	), ImageConvert)
	defaultRegistry.Register("image", mcp.NewTool("image_rotate",
		mcp.WithDescription("Rotate an image by 90, 180, or 270 degrees clockwise"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Image path")),
		mcp.WithNumber("degrees", mcp.Required(), mcp.Description("Rotation in degrees: 90, 180, or 270")),
		mcp.WithString("output", mcp.Description("Optional output path")),
	), ImageRotate)
	defaultRegistry.Register("image", mcp.NewTool("image_grayscale",
		mcp.WithDescription("Convert an image to grayscale"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Image path")),
		mcp.WithString("output", mcp.Description("Optional output path")),
	), ImageGrayscale)
	defaultRegistry.Register("image", mcp.NewTool("image_flip",
		mcp.WithDescription("Flip an image horizontally or vertically"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Image path")),
		mcp.WithString("direction", mcp.Required(), mcp.Description("Flip direction: horizontal or vertical")),
		mcp.WithString("output", mcp.Description("Optional output path")),
	), ImageFlip)

	defaultRegistry.Register("data", mcp.NewTool("data_json_format",
		mcp.WithDescription("Pretty-print JSON"),
		mcp.WithString("json", mcp.Required(), mcp.Description("JSON input")),
	), DataJSONFormat)
	defaultRegistry.Register("data", mcp.NewTool("data_csv_to_json",
		mcp.WithDescription("Convert CSV to JSON"),
		mcp.WithString("csv", mcp.Required(), mcp.Description("CSV input")),
	), DataCSVToJSON)
	defaultRegistry.Register("data", mcp.NewTool("data_json_to_csv",
		mcp.WithDescription("Convert JSON to CSV"),
		mcp.WithString("json", mcp.Required(), mcp.Description("JSON array")),
	), DataJSONToCSV)
	defaultRegistry.Register("data", mcp.NewTool("data_json_query",
		mcp.WithDescription("Query JSON by path"),
		mcp.WithString("json", mcp.Required(), mcp.Description("JSON input")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Dot path")),
	), DataJSONQuery)

	defaultRegistry.Register("text", mcp.NewTool("text_word_count",
		mcp.WithDescription("Count words in text"),
		mcp.WithString("text", mcp.Required(), mcp.Description("Input text")),
	), TextWordCount)
	defaultRegistry.Register("text", mcp.NewTool("text_replace",
		mcp.WithDescription("Replace text content"),
		mcp.WithString("text", mcp.Required(), mcp.Description("Input text")),
		mcp.WithString("find", mcp.Required(), mcp.Description("Text to find")),
		mcp.WithString("replace", mcp.Required(), mcp.Description("Replacement text")),
		mcp.WithBoolean("all", mcp.Description("Replace all matches")),
	), TextReplace)
	defaultRegistry.Register("text", mcp.NewTool("text_extract_regex",
		mcp.WithDescription("Extract regex matches"),
		mcp.WithString("text", mcp.Required(), mcp.Description("Input text")),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("Regex pattern")),
	), TextExtractRegex)
	defaultRegistry.Register("text", mcp.NewTool("text_base64_encode",
		mcp.WithDescription("Encode text as base64"),
		mcp.WithString("text", mcp.Required(), mcp.Description("Input text")),
	), TextBase64Encode)
	defaultRegistry.Register("text", mcp.NewTool("text_base64_decode",
		mcp.WithDescription("Decode base64 text"),
		mcp.WithString("encoded", mcp.Required(), mcp.Description("Base64 text")),
	), TextBase64Decode)
	defaultRegistry.Register("text", mcp.NewTool("text_md_to_html",
		mcp.WithDescription("Convert markdown to HTML"),
		mcp.WithString("markdown", mcp.Required(), mcp.Description("Markdown text")),
	), TextMarkdownToHTML)

	defaultRegistry.Register("web", mcp.NewTool("web_fetch",
		mcp.WithDescription("Fetch a URL body"),
		mcp.WithString("url", mcp.Required(), mcp.Description("URL to fetch")),
		mcp.WithString("method", mcp.Description("HTTP method")),
		mcp.WithObject("headers", mcp.Description("Request headers")),
	), WebFetch)
	defaultRegistry.Register("web", mcp.NewTool("web_status",
		mcp.WithDescription("Check URL reachability"),
		mcp.WithString("url", mcp.Required(), mcp.Description("URL to check")),
	), WebStatus)
}
