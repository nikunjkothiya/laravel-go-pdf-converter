package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/signintech/gopdf"
)

// Builder provides a fluent interface for creating PDF documents
type Builder struct {
	pdf       *gopdf.GoPdf
	options   Options
	currentY  float64
	pageNum   int
	fontLoaded bool
}

// NewBuilder creates a new PDF builder with the given options
func NewBuilder(opts Options) (*Builder, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *opts.GetPageRect()})

	b := &Builder{
		pdf:      pdf,
		options:  opts,
		currentY: opts.Margin,
		pageNum:  0,
	}

	// Load default font
	if err := b.loadFont(); err != nil {
		return nil, err
	}

	return b, nil
}

// loadFont loads the specified font or falls back to built-in
func (b *Builder) loadFont() error {
	// Try to use system fonts first
	fontPaths := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"/System/Library/Fonts/Helvetica.ttc",
		"C:\\Windows\\Fonts\\arial.ttf",
	}

	for _, fontPath := range fontPaths {
		if _, err := os.Stat(fontPath); err == nil {
			if err := b.pdf.AddTTFFont("default", fontPath); err == nil {
				b.fontLoaded = true
				return b.pdf.SetFont("default", "", b.options.FontSize)
			}
		}
	}

	// Fall back to built-in font (limited Unicode support)
	// gopdf doesn't have built-in fonts, so we embed a minimal font
	return b.embedMinimalFont()
}

// embedMinimalFont embeds a minimal font for basic operation
func (b *Builder) embedMinimalFont() error {
	// Use the TTF font embedded in the binary or from a known location
	// For now, we'll try common locations
	homeDir, _ := os.UserHomeDir()
	additionalPaths := []string{
		filepath.Join(homeDir, ".fonts", "DejaVuSans.ttf"),
		"/usr/local/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	}

	for _, fontPath := range additionalPaths {
		if _, err := os.Stat(fontPath); err == nil {
			if err := b.pdf.AddTTFFont("default", fontPath); err == nil {
				b.fontLoaded = true
				return b.pdf.SetFont("default", "", b.options.FontSize)
			}
		}
	}

	return nil // Proceed without font, will use basic rendering
}

// AddPage adds a new page to the document
func (b *Builder) AddPage() {
	b.pdf.AddPage()
	b.currentY = b.options.Margin
	b.pageNum++
	
	// Draw global header and footer
	b.drawHeader()
	b.drawFooter()
	
	// Reset Y to below header
	b.currentY = b.options.Margin + 20
}

func (b *Builder) drawHeader() {
	if b.options.HeaderText == "" {
		return
	}
	
	style := DefaultStyle()
	style.FontSize = 8
	style.TextColor = ColorGray
	
	b.SetFont(style.FontFamily, style.FontStyle, style.FontSize)
	b.SetTextColor(style.TextColor)
	
	// Draw header text at the top
	b.pdf.SetX(b.options.Margin)
	b.pdf.SetY(b.options.Margin - 10)
	b.pdf.Cell(nil, b.options.HeaderText)
}

func (b *Builder) drawFooter() {
	pageHeight := b.options.PageSize.Height
	if b.options.Orientation == Landscape {
		pageHeight = b.options.PageSize.Width
	}
	
	style := DefaultStyle()
	style.FontSize = 8
	style.TextColor = ColorGray
	
	b.SetFont(style.FontFamily, style.FontStyle, style.FontSize)
	b.SetTextColor(style.TextColor)
	
	// Draw footer text (custom text + Page X of Y)
	footerY := pageHeight - b.options.Margin + 5
	
	// Custom footer text (left aligned)
	if b.options.FooterText != "" {
		b.pdf.SetX(b.options.Margin)
		b.pdf.SetY(footerY)
		b.pdf.Cell(nil, b.options.FooterText)
	}
	
	// Page X of Y (right aligned)
	pageStr := fmt.Sprintf("Page %d of ", b.pageNum)
	textWidth, _ := b.pdf.MeasureTextWidth(pageStr)
	placeholderWidth := 20.0
	totalWidth := textWidth + placeholderWidth
	
	b.pdf.SetX(b.options.ContentWidth() + b.options.Margin - totalWidth)
	b.pdf.SetY(footerY)
	b.pdf.Cell(nil, pageStr)
	b.pdf.PlaceHolderText("total", placeholderWidth)
}

// SetFont sets the current font
func (b *Builder) SetFont(family string, style string, size float64) error {
	if !b.fontLoaded {
		return nil
	}
	return b.pdf.SetFont("default", style, size)
}

// SetTextColor sets the text color
func (b *Builder) SetTextColor(c Color) {
	b.pdf.SetTextColor(c.R, c.G, c.B)
}

// SetFillColor sets the fill color
func (b *Builder) SetFillColor(c Color) {
	b.pdf.SetFillColor(c.R, c.G, c.B)
}

// SetStrokeColor sets the stroke/border color
func (b *Builder) SetStrokeColor(c Color) {
	b.pdf.SetStrokeColor(c.R, c.G, c.B)
}

// GetX returns current X position
func (b *Builder) GetX() float64 {
	return b.pdf.GetX()
}

// GetY returns current Y position
func (b *Builder) GetY() float64 {
	return b.currentY
}

// SetXY sets the current position
func (b *Builder) SetXY(x, y float64) {
	b.pdf.SetX(x)
	b.pdf.SetY(y)
	b.currentY = y
}

// Cell draws a cell with text
func (b *Builder) Cell(w, h float64, text string, style Style) error {
	x := b.pdf.GetX()
	y := b.currentY

	// Draw background if specified
	if style.HasBackground {
		b.SetFillColor(style.FillColor)
		b.pdf.Rectangle(x, y, x+w, y+h, "F", 0, 0)
	}

	// Draw border if specified
	if style.HasBorder {
		b.SetStrokeColor(style.BorderColor)
		b.pdf.SetLineWidth(style.BorderWidth)
		b.pdf.Rectangle(x, y, x+w, y+h, "D", 0, 0)
	}

	// Draw text with padding and alignment
	b.SetTextColor(style.TextColor)
	
	textWidth := b.MeasureTextWidth(text)
	maxWidth := w - (style.Padding * 2)
	if textWidth > maxWidth {
		text = b.truncateText(text, maxWidth)
		textWidth = b.MeasureTextWidth(text)
	}

	var textX float64
	switch style.Alignment {
	case AlignCenter:
		textX = x + (w-textWidth)/2
	case AlignRight:
		textX = x + w - textWidth - style.Padding
	default: // AlignLeft
		textX = x + style.Padding
	}

	textY := y + style.Padding + style.FontSize

	b.pdf.SetX(textX)
	b.pdf.SetY(textY)
	b.pdf.Text(text)

	// Move to next cell position
	b.pdf.SetX(x + w)

	return nil
}

// truncateText truncates text to fit within maxWidth
func (b *Builder) truncateText(text string, maxWidth float64) string {
	if !b.fontLoaded {
		// Rough estimate: 6 points per character
		maxChars := int(maxWidth / 6)
		if len(text) > maxChars && maxChars > 3 {
			return text[:maxChars-3] + "..."
		}
		return text
	}

	width, _ := b.pdf.MeasureTextWidth(text)
	if width <= maxWidth {
		return text
	}

	// Binary search for the right length
	for len(text) > 3 {
		text = text[:len(text)-1]
		width, _ = b.pdf.MeasureTextWidth(text + "...")
		if width <= maxWidth {
			return text + "..."
		}
	}

	return text
}

// MeasureTextWidth measures the width of text
func (b *Builder) MeasureTextWidth(text string) float64 {
	if !b.fontLoaded {
		return float64(len(text)) * 6 // Rough estimate
	}
	width, _ := b.pdf.MeasureTextWidth(text)
	return width
}

// NewLine moves to a new line
func (b *Builder) NewLine(height float64) {
	b.currentY += height
	b.pdf.SetX(b.options.Margin)
	b.pdf.SetY(b.currentY)
}

// NewLineAt moves to a new line and sets X to startX
func (b *Builder) NewLineAt(height float64, startX float64) {
	b.currentY += height
	b.pdf.SetX(startX)
	b.pdf.SetY(b.currentY)
}

// NeedsNewPage checks if we need a new page for the given height
func (b *Builder) NeedsNewPage(height float64) bool {
	pageHeight := b.options.PageSize.Height
	if b.options.Orientation == Landscape {
		pageHeight = b.options.PageSize.Width
	}
	return b.currentY+height > pageHeight-b.options.Margin
}

// DrawTable draws a complete table from data
func (b *Builder) DrawTable(headers []string, rows [][]string, colWidths []float64) error {
	style := DefaultStyle()
	headerStyle := HeaderStyle()
	rowHeight := style.FontSize + (style.Padding * 2) + 4

	// Calculate total table width for centering
	tableWidth := 0.0
	for _, w := range colWidths {
		tableWidth += w
	}
	
	startX := b.options.Margin
	contentWidth := b.options.ContentWidth()
	if tableWidth < contentWidth {
		startX = b.options.Margin + (contentWidth-tableWidth)/2
	}

	// Draw headers
	if len(headers) > 0 && b.options.HeaderRow {
		b.SetFont(headerStyle.FontFamily, headerStyle.FontStyle, headerStyle.FontSize)
		b.pdf.SetX(startX)

		for i, header := range headers {
			if i < len(colWidths) {
				if err := b.Cell(colWidths[i], rowHeight, header, headerStyle); err != nil {
					return err
				}
			}
		}
		b.NewLineAt(rowHeight, startX)
	}

	// Draw data rows
	b.SetFont(style.FontFamily, style.FontStyle, style.FontSize)
	for rowIdx, row := range rows {
		// Check if we need a new page
		if b.NeedsNewPage(rowHeight) {
			b.AddPage()
			// Redraw headers on new page
			if b.options.HeaderRow && len(headers) > 0 {
				b.SetFont(headerStyle.FontFamily, headerStyle.FontStyle, headerStyle.FontSize)
				b.pdf.SetX(startX)
				for i, header := range headers {
					if i < len(colWidths) {
						b.Cell(colWidths[i], rowHeight, header, headerStyle)
					}
				}
				b.NewLineAt(rowHeight, startX)
				b.SetFont(style.FontFamily, style.FontStyle, style.FontSize)
			}
		}

		rowStyle := TableStyle()
		if rowIdx%2 == 1 {
			rowStyle.FillColor = ColorLightGray
			rowStyle.HasBackground = true
		}
		
		b.pdf.SetX(startX)

		for i, cell := range row {
			if i < len(colWidths) {
				// Detect alignment based on content (simple heuristic)
				cellStyle := rowStyle
				if isNumeric(cell) {
					cellStyle.Alignment = AlignRight
				}
 				
				if err := b.Cell(colWidths[i], rowHeight, cell, cellStyle); err != nil {
					return err
				}
			}
		}
		b.NewLineAt(rowHeight, startX)
	}

	return nil
}

// AddText adds a text paragraph
func (b *Builder) AddText(text string, style Style) error {
	b.SetFont(style.FontFamily, style.FontStyle, style.FontSize)
	b.SetTextColor(style.TextColor)
	b.pdf.SetX(b.options.Margin)
	b.pdf.SetY(b.currentY)
	b.pdf.Text(text)
	b.NewLine(style.FontSize * style.LineHeight)
	return nil
}

// AddImage adds an image from file
func (b *Builder) AddImage(imagePath string, x, y, w, h float64) error {
	return b.pdf.Image(imagePath, x, y, &gopdf.Rect{W: w, H: h})
}

// Save writes the PDF to the specified path
func (b *Builder) Save(outputPath string) error {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Fill in total page count
	b.pdf.FillInPlaceHoldText("total", fmt.Sprintf("%d", b.pageNum), gopdf.Left)
	
	return b.pdf.WritePdf(outputPath)
}

// isNumeric checks if a string represents a number
func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	// Simple check for digits, decimal point, and signs
	hasDigit := false
	for _, r := range s {
		if r >= '0' && r <= '9' {
			hasDigit = true
			continue
		}
		if r == '.' || r == '-' || r == '+' || r == ',' {
			continue
		}
		return false
	}
	return hasDigit
}

// GetPdf returns the underlying GoPdf instance for advanced operations
func (b *Builder) GetPdf() *gopdf.GoPdf {
	return b.pdf
}
