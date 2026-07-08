import { Box, Text, useState, SelectInput, useInput, html } from "../h.js";
import { GradientRule, KeyHint, Divider } from "../components.js";
const TOOLS = {
  "File (7)": [
    { name: "file_read", description: "Read file contents as text" },
    { name: "file_write", description: "Write text to a file" },
    { name: "file_list", description: "List files in a directory" },
    { name: "file_size", description: "Get file size in bytes" },
    { name: "file_hash", description: "Compute SHA-256 hash of a file" },
    { name: "file_zip", description: "Create a zip archive from files" },
    { name: "file_unzip", description: "Extract files from a zip archive" },
  ],
  "Image (7)": [
    { name: "image_info", description: "Get image dimensions, format, size" },
    { name: "image_resize", description: "Resize an image to target dimensions" },
    { name: "image_crop", description: "Crop an image to a region" },
    { name: "image_convert", description: "Convert image format (png, jpg, webp)" },
    { name: "image_rotate", description: "Rotate an image by degrees" },
    { name: "image_grayscale", description: "Convert image to grayscale" },
    { name: "image_flip", description: "Flip image horizontally or vertically" },
  ],
  "PDF (6)": [
    { name: "pdf_info", description: "Get PDF metadata and page count" },
    { name: "pdf_page_count", description: "Count pages in a PDF" },
    { name: "pdf_extract_text", description: "Extract text from PDF pages" },
    { name: "pdf_extract_pages", description: "Extract specific pages as new PDF" },
    { name: "pdf_merge", description: "Merge multiple PDFs into one" },
    { name: "pdf_split", description: "Split a PDF into individual pages" },
  ],
  "Text (6)": [
    { name: "text_word_count", description: "Count words, lines, characters" },
    { name: "text_replace", description: "Find and replace text" },
    { name: "text_extract_regex", description: "Extract matches using regex" },
    { name: "text_base64_encode", description: "Encode text to base64" },
    { name: "text_base64_decode", description: "Decode base64 to text" },
    { name: "text_md_to_html", description: "Convert markdown to HTML" },
  ],
  "Data (4)": [
    { name: "data_csv_to_json", description: "Convert CSV to JSON array" },
    { name: "data_json_to_csv", description: "Convert JSON array to CSV" },
    { name: "data_json_format", description: "Format/pretty-print JSON" },
    { name: "data_json_query", description: "Query JSON with JMESPath" },
  ],
  "Web (2)": [
    { name: "web_fetch", description: "Fetch a URL and return content" },
    { name: "web_status", description: "Check HTTP status code of a URL" },
  ],
  "Workflow (8)": [
    { name: "run_workflow", description: "Execute a full 5-phase workflow" },
    { name: "pm_plan", description: "Phase 1: PM planning" },
    { name: "run_research", description: "Phase 2: Research" },
    { name: "run_brand", description: "Phase 3: Brand" },
    { name: "run_ux", description: "Phase 4: UX" },
    { name: "run_gtm", description: "Phase 5: Go-to-market" },
    { name: "approval", description: "Human checkpoint between phases" },
    { name: "assemble_plan", description: "Assemble final plan from phases" },
  ],
};
export function ToolsBrowser({ onDone }) {
  const [view, setView] = useState("categories"); 
  const [selectedCategory, setSelectedCategory] = useState(null);
  const [selectedTool, setSelectedTool] = useState(null);
  useInput((input, key) => {
    if (key.escape) {
      if (view === "detail") {
        setView("tools");
      } else if (view === "tools") {
        setView("categories");
      } else if (view === "categories") {
        onDone();
      }
    }
  });
  if (view === "categories") {
    const categories = Object.keys(TOOLS);
    const items = categories.map((cat) => ({
      label: `${cat}  (${TOOLS[cat].length} tools)`,
      value: cat,
    }));
    return html`
      <${Box} flexDirection="column" paddingTop=${1}>
        <${Text} color="cyan" bold>Tool Categories<//>
        <${GradientRule} width=${36} />
        <${Box} marginTop=${1}>
          <${SelectInput} items=${items} onSelect=${(item) => {
            setSelectedCategory(item.value);
            setView("tools");
          }} />
        <//>
        <${KeyHint} hints=${["↑↓ navigate", "Enter select", "Esc back"]} />
      <//>
    `;
  }
  if (view === "tools") {
    const tools = TOOLS[selectedCategory];
    const items = tools.map((t) => ({
      label: `${t.name}  —  ${t.description}`,
      value: t.name,
    }));
    return html`
      <${Box} flexDirection="column" paddingTop=${1}>
        <${Text} color="gray}> Categories<//>
        <${Text} color="cyan" bold>${selectedCategory}<//>
        <${GradientRule} width=${36} />
        <${Box} marginTop=${1}>
          <${SelectInput} items=${items} onSelect=${(item) => {
            const tool = tools.find((t) => t.name === item.value);
            setSelectedTool(tool);
            setView("detail");
          }} />
        <//>
        <${KeyHint} hints=${["↑↓ navigate", "Enter detail", "Esc back"]} />
      <//>
    `;
  }
  if (view === "detail") {
    return html`
      <${Box} flexDirection="column" paddingTop=${2} paddingLeft=${2}>
        <${Text} color="gray">> ${selectedCategory} > ${selectedTool.name}<//>
        <${Text} color="cyan" bold>${selectedTool.name}<//>
        <${Divider} width=${40} />
        <${Box} marginTop=${1}>
          <${Text} color="white">${selectedTool.description}<//>
        <//>
        <${Box} marginTop=${1}>
          <${Text} color="gray">Category: ${selectedCategory}<//>
        <//>
        <${Box} marginTop=${2}>
          <${SelectInput}
            items=${[
              { label: "Back to category", value: "cat" },
              { label: "Back to all categories", value: "all" },
            ]}
            onSelect=${(item) => {
              if (item.value === "cat") setView("tools");
              else setView("categories");
            }}
          />
        <//>
        <${KeyHint} hints=${["Esc back"]} />
      <//>
    `;
  }
}
