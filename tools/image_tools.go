package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func ImageInfo(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	img, format, err := loadImage(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("load image: %v", err)), nil
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("stat image: %v", err)), nil
	}

	bounds := img.Bounds()
	return imageJSONResult(map[string]any{
		"width":           bounds.Dx(),
		"height":          bounds.Dy(),
		"format":          format,
		"file_size_bytes": info.Size(),
	}), nil
}

func ImageResize(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	img, format, err := loadImage(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("load image: %v", err)), nil
	}

	srcBounds := img.Bounds()
	srcW, srcH := srcBounds.Dx(), srcBounds.Dy()
	newW := req.GetInt("width", 0)
	newH := req.GetInt("height", 0)
	if newW <= 0 && newH <= 0 {
		return mcp.NewToolResultError("width or height is required"), nil
	}
	if newW <= 0 {
		newW = maxImageInt(1, int(math.Round(float64(srcW)*float64(newH)/float64(srcH))))
	}
	if newH <= 0 {
		newH = maxImageInt(1, int(math.Round(float64(srcH)*float64(newW)/float64(srcW))))
	}
	if newW <= 0 || newH <= 0 {
		return mcp.NewToolResultError("width and height must be positive"), nil
	}

	output, err := resolveOptionalOutput(req.GetString("output", ""), outputPath(resolved, "_resized", ""))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := saveImage(resizeNN(img, newW, newH), output, format, 85); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("save image: %v", err)), nil
	}

	return imageJSONResult(map[string]any{
		"output_path": output,
		"width":       newW,
		"height":      newH,
	}), nil
}

func ImageCrop(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	x, err := requiredIntArg(req, "x")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	y, err := requiredIntArg(req, "y")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	width, err := requiredIntArg(req, "width")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	height, err := requiredIntArg(req, "height")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if width <= 0 || height <= 0 {
		return mcp.NewToolResultError("width and height must be positive"), nil
	}

	img, format, err := loadImage(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("load image: %v", err)), nil
	}

	bounds := img.Bounds()
	if x < bounds.Min.X || y < bounds.Min.Y || x+width > bounds.Max.X || y+height > bounds.Max.Y {
		return mcp.NewToolResultError("crop rectangle is outside image bounds"), nil
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, dst.Bounds(), img, image.Point{X: x, Y: y}, draw.Src)

	output, err := resolveOptionalOutput(req.GetString("output", ""), outputPath(resolved, "_cropped", ""))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := saveImage(dst, output, format, 85); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("save image: %v", err)), nil
	}

	return imageJSONResult(map[string]any{
		"output_path": output,
		"width":       width,
		"height":      height,
	}), nil
}

func ImageConvert(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	targetFormat := normalizeImageFormat(req.GetString("format", ""))
	if targetFormat == "" {
		return mcp.NewToolResultError("format is required"), nil
	}
	if !isSupportedImageFormat(targetFormat) {
		return mcp.NewToolResultError("format must be png, jpeg, or gif"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	img, _, err := loadImage(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("load image: %v", err)), nil
	}

	quality := req.GetInt("quality", 85)
	if targetFormat == "jpeg" && (quality < 1 || quality > 100) {
		return mcp.NewToolResultError("quality must be between 1 and 100"), nil
	}

	output := outputPath(resolved, "", formatExtension(targetFormat))
	if err := saveImage(img, output, targetFormat, quality); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("save image: %v", err)), nil
	}

	return imageJSONResult(map[string]any{
		"output_path": output,
	}), nil
}

func ImageRotate(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	degrees, err := requiredIntArg(req, "degrees")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	img, format, err := loadImage(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("load image: %v", err)), nil
	}

	var rotated image.Image
	switch degrees {
	case 90:
		rotated = rotate90(img)
	case 180:
		rotated = rotate180(img)
	case 270:
		rotated = rotate270(img)
	default:
		return mcp.NewToolResultError("degrees must be 90, 180, or 270"), nil
	}

	output, err := resolveOptionalOutput(req.GetString("output", ""), outputPath(resolved, fmt.Sprintf("_rotated_%d", degrees), ""))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := saveImage(rotated, output, format, 85); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("save image: %v", err)), nil
	}

	return imageJSONResult(map[string]any{
		"output_path": output,
	}), nil
}

func ImageGrayscale(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	img, format, err := loadImage(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("load image: %v", err)), nil
	}

	bounds := img.Bounds()
	dst := image.NewGray(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dst.Set(x, y, color.GrayModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+y)))
		}
	}

	output, err := resolveOptionalOutput(req.GetString("output", ""), outputPath(resolved, "_grayscale", ""))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := saveImage(dst, output, format, 85); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("save image: %v", err)), nil
	}

	return imageJSONResult(map[string]any{
		"output_path": output,
	}), nil
}

func ImageFlip(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	direction := strings.ToLower(strings.TrimSpace(req.GetString("direction", "")))
	if direction == "" {
		return mcp.NewToolResultError("direction is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	img, format, err := loadImage(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("load image: %v", err)), nil
	}

	var flipped image.Image
	switch direction {
	case "horizontal":
		flipped = flipHorizontal(img)
	case "vertical":
		flipped = flipVertical(img)
	default:
		return mcp.NewToolResultError("direction must be horizontal or vertical"), nil
	}

	output, err := resolveOptionalOutput(req.GetString("output", ""), outputPath(resolved, "_flipped_"+direction, ""))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := saveImage(flipped, output, format, 85); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("save image: %v", err)), nil
	}

	return imageJSONResult(map[string]any{
		"output_path": output,
	}), nil
}

func loadImage(path string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}
	return img, normalizeImageFormat(format), nil
}

func saveImage(img image.Image, path string, format string, quality int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch normalizeImageFormat(format) {
	case "png":
		return png.Encode(file, img)
	case "jpeg":
		if quality <= 0 {
			quality = 85
		}
		return jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
	case "gif":
		return gif.Encode(file, img, nil)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
}

func outputPath(input, suffix, newExt string) string {
	ext := filepath.Ext(input)
	base := strings.TrimSuffix(input, ext)
	if newExt == "" {
		newExt = ext
	}
	if newExt != "" && !strings.HasPrefix(newExt, ".") {
		newExt = "." + newExt
	}
	return base + suffix + newExt
}

func resizeNN(src image.Image, newW, newH int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := x*srcW/newW + srcBounds.Min.X
			srcY := y*srcH/newH + srcBounds.Min.Y
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func rotate90(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(h-1-y, x, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst
}

func rotate180(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(w-1-x, h-1-y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst
}

func rotate270(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(y, w-1-x, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst
}

func flipHorizontal(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(w-1-x, y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst
}

func flipVertical(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, h-1-y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst
}

func imageJSONResult(v any) *mcp.CallToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err))
	}
	return mcp.NewToolResultText(string(data))
}

func requiredIntArg(req mcp.CallToolRequest, key string) (int, error) {
	args := req.GetArguments()
	val, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("%s is required", key)
	}
	switch typed := val.(type) {
	case int:
		return typed, nil
	case float64:
		return int(typed), nil
	case string:
		return req.RequireInt(key)
	default:
		return 0, fmt.Errorf("%s must be an integer", key)
	}
}

func resolveOptionalOutput(rawOutput, defaultPath string) (string, error) {
	if strings.TrimSpace(rawOutput) == "" {
		return defaultPath, nil
	}
	return resolveToolPath(rawOutput)
}

func normalizeImageFormat(format string) string {
	format = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(format, ".")))
	if format == "jpg" {
		return "jpeg"
	}
	return format
}

func isSupportedImageFormat(format string) bool {
	switch normalizeImageFormat(format) {
	case "png", "jpeg", "gif":
		return true
	default:
		return false
	}
}

func formatExtension(format string) string {
	switch normalizeImageFormat(format) {
	case "jpeg":
		return ".jpeg"
	case "png":
		return ".png"
	case "gif":
		return ".gif"
	default:
		return ""
	}
}

func maxImageInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
