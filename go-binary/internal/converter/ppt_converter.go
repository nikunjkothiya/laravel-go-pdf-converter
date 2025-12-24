package converter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf16"

	"github.com/nikunjkothiya/gopdfconv/internal/pdf"
	"github.com/nikunjkothiya/gopdfconv/pkg/errors"
	"github.com/richardlehane/mscfb"
)

// PPTConverter handles legacy PowerPoint (.ppt) to PDF conversion
// Uses OLE compound document parsing for text extraction
type PPTConverter struct {
	opts pdf.Options
}

// NewPPTConverter creates a new PPT converter
func NewPPTConverter() *PPTConverter {
	return &PPTConverter{
		opts: pdf.DefaultOptions(),
	}
}

// SupportedExtensions returns extensions handled by this converter
func (c *PPTConverter) SupportedExtensions() []string {
	return []string{".ppt"}
}

// Validate checks if the input file is a valid PPT file
func (c *PPTConverter) Validate(inputPath string) error {
	// Check file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return errors.NewWithFile(errors.ErrFileNotFound, "File not found", inputPath)
	}

	// Try to open as OLE compound document
	file, err := os.Open(inputPath)
	if err != nil {
		return errors.NewWithFile(errors.ErrFileNotFound, "Cannot open file", inputPath)
	}
	defer file.Close()

	doc, err := mscfb.New(file)
	if err != nil {
		return errors.NewWithDetails(errors.ErrInvalidFormat, "Not a valid PPT file (invalid OLE format)", inputPath, err.Error())
	}

	// Check for PowerPoint Document stream
	hasPPTStream := false
	for entry, err := doc.Next(); err == nil; entry, err = doc.Next() {
		if entry.Name == "PowerPoint Document" || entry.Name == "Current User" {
			hasPPTStream = true
			break
		}
	}

	if !hasPPTStream {
		return errors.NewWithFile(errors.ErrInvalidFormat, "Not a valid PPT file (missing PowerPoint streams)", inputPath)
	}

	return nil
}

// Convert performs the PPT to PDF conversion
func (c *PPTConverter) Convert(inputPath, outputPath string, opts pdf.Options) error {
	// Validate input
	if err := c.Validate(inputPath); err != nil {
		return err
	}

	// Open file
	file, err := os.Open(inputPath)
	if err != nil {
		return errors.NewWithFile(errors.ErrFileNotFound, "Cannot open file", inputPath)
	}
	defer file.Close()

	// Parse OLE compound document
	doc, err := mscfb.New(file)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to parse PPT file")
	}

	// Extract text from PowerPoint Document stream
	slides, err := c.extractSlides(doc)
	if err != nil {
		return err
	}

	// Create PDF
	builder, err := pdf.NewBuilder(opts)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create PDF builder")
	}

	// Render slides
	c.renderSlides(builder, slides, opts)

	// Save PDF
	if err := builder.Save(outputPath); err != nil {
		return errors.Wrap(err, errors.ErrWriteFailed, "Failed to save PDF")
	}

	return nil
}

// PPTSlide represents extracted slide content
type PPTSlide struct {
	Index int
	Title string
	Body  []string
}

// extractSlides extracts text content from PPT file
func (c *PPTConverter) extractSlides(doc *mscfb.Reader) ([]PPTSlide, error) {
	var slides []PPTSlide
	var pptData []byte

	// Find and read PowerPoint Document stream
	for entry, err := doc.Next(); err == nil; entry, err = doc.Next() {
		if entry.Name == "PowerPoint Document" {
			pptData, err = io.ReadAll(entry)
			if err != nil {
				return nil, errors.Wrap(err, errors.ErrConversionFailed, "Failed to read PowerPoint Document stream")
			}
			break
		}
	}

	if len(pptData) == 0 {
		return nil, errors.New(errors.ErrInvalidFormat, "No PowerPoint Document stream found")
	}

	// Parse PPT binary format to extract text
	texts := c.extractTextFromPPTBinary(pptData)

	// Group texts into slides (rough heuristic)
	currentSlide := PPTSlide{Index: 1}
	for i, text := range texts {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		// First non-empty text on slide is title
		if currentSlide.Title == "" {
			currentSlide.Title = text
		} else {
			currentSlide.Body = append(currentSlide.Body, text)
		}

		// Heuristic: new slide every 5-7 text blocks or on specific markers
		if len(currentSlide.Body) >= 6 || (i > 0 && len(text) > 50 && strings.HasSuffix(text, ".")) {
			slides = append(slides, currentSlide)
			currentSlide = PPTSlide{Index: len(slides) + 1}
		}
	}

	// Add last slide if has content
	if currentSlide.Title != "" || len(currentSlide.Body) > 0 {
		slides = append(slides, currentSlide)
	}

	// If no slides parsed, create one with all text
	if len(slides) == 0 && len(texts) > 0 {
		slides = append(slides, PPTSlide{
			Index: 1,
			Title: "Slide Content",
			Body:  texts,
		})
	}

	return slides, nil
}

// extractTextFromPPTBinary extracts text strings from PPT binary data
func (c *PPTConverter) extractTextFromPPTBinary(data []byte) []string {
	var texts []string

	// PPT files contain text in various record types
	// We look for TextCharsAtom (0x0FA0) and TextBytesAtom (0x0FA8)
	// Also look for plain text patterns

	// Method 1: Look for Unicode text patterns
	unicodeTexts := c.extractUnicodeStrings(data)
	texts = append(texts, unicodeTexts...)

	// Method 2: Look for ASCII text patterns
	asciiTexts := c.extractASCIIStrings(data)
	texts = append(texts, asciiTexts...)

	// Remove duplicates and very short strings
	seen := make(map[string]bool)
	var uniqueTexts []string
	for _, t := range texts {
		t = strings.TrimSpace(t)
		if len(t) < 3 || seen[t] {
			continue
		}
		// Filter out binary garbage
		if c.isValidText(t) {
			seen[t] = true
			uniqueTexts = append(uniqueTexts, t)
		}
	}

	return uniqueTexts
}

// extractUnicodeStrings extracts UTF-16LE strings from binary data
func (c *PPTConverter) extractUnicodeStrings(data []byte) []string {
	var texts []string

	// Look for TextCharsAtom records (type 0x0FA0)
	for i := 0; i < len(data)-8; i++ {
		// Check for record header pattern
		if i+4 < len(data) {
			recType := binary.LittleEndian.Uint16(data[i : i+2])
			
			// TextCharsAtom = 0x0FA0 (4000)
			if recType == 0x0FA0 {
				if i+8 < len(data) {
					recLen := binary.LittleEndian.Uint32(data[i+4 : i+8])
					if recLen > 0 && recLen < 10000 && int(i+8+int(recLen)) <= len(data) {
						textData := data[i+8 : i+8+int(recLen)]
						text := c.decodeUTF16LE(textData)
						if len(text) > 2 {
							texts = append(texts, text)
						}
						i += 8 + int(recLen) - 1
					}
				}
			}
		}
	}

	return texts
}

// decodeUTF16LE decodes UTF-16LE bytes to string
func (c *PPTConverter) decodeUTF16LE(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	// Convert bytes to uint16 slice
	u16s := make([]uint16, len(data)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = binary.LittleEndian.Uint16(data[i*2:])
	}

	// Decode UTF-16 to string
	runes := utf16.Decode(u16s)
	return string(runes)
}

// extractASCIIStrings extracts printable ASCII strings
func (c *PPTConverter) extractASCIIStrings(data []byte) []string {
	var texts []string
	var current bytes.Buffer

	for _, b := range data {
		// Printable ASCII range
		if b >= 32 && b < 127 {
			current.WriteByte(b)
		} else if current.Len() > 0 {
			if current.Len() >= 4 {
				texts = append(texts, current.String())
			}
			current.Reset()
		}
	}

	if current.Len() >= 4 {
		texts = append(texts, current.String())
	}

	return texts
}

// isValidText checks if a string looks like real text (not binary garbage)
func (c *PPTConverter) isValidText(s string) bool {
	if len(s) < 3 {
		return false
	}

	// Count valid characters
	validCount := 0
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
		   (r >= '0' && r <= '9') || r == ' ' || r == '.' || 
		   r == ',' || r == ':' || r == '-' || r == '!' || r == '?' {
			validCount++
		}
	}

	// At least 60% should be valid text characters
	return float64(validCount)/float64(len(s)) > 0.6
}

// renderSlides renders extracted slides to PDF
func (c *PPTConverter) renderSlides(builder *pdf.Builder, slides []PPTSlide, opts pdf.Options) {
	titleStyle := pdf.HeaderStyle()
	titleStyle.FontSize = 24

	bodyStyle := pdf.DefaultStyle()
	bodyStyle.FontSize = 12

	noteStyle := pdf.DefaultStyle()
	noteStyle.FontSize = 10
	noteStyle.TextColor = pdf.ColorGray

	for i, slide := range slides {
		if i > 0 {
			builder.AddPage()
		} else {
			builder.AddPage()
		}

		// Add slide number
		slideNumStyle := pdf.DefaultStyle()
		slideNumStyle.FontSize = 8
		slideNumStyle.TextColor = pdf.ColorGray

		// Render title
		if slide.Title != "" {
			builder.AddText(slide.Title, titleStyle)
			builder.NewLine(20)
		}

		// Render body text
		for _, text := range slide.Body {
			builder.AddText("• "+text, bodyStyle)
			builder.NewLine(bodyStyle.FontSize + 6)
		}

		// Add slide number at bottom
		builder.SetXY(opts.ContentWidth()-30, opts.ContentHeight()-10)
		builder.AddText(fmt.Sprintf("Slide %d", slide.Index), slideNumStyle)
	}

	// If no slides, add a note
	if len(slides) == 0 {
		builder.AddPage()
		builder.AddText("No text content could be extracted from this PPT file.", noteStyle)
		builder.NewLine(20)
		builder.AddText("For best results with legacy .ppt files, consider:", noteStyle)
		builder.NewLine(noteStyle.FontSize + 4)
		builder.AddText("1. Converting to .pptx format first", noteStyle)
		builder.NewLine(noteStyle.FontSize + 4)
		builder.AddText("2. Installing LibreOffice for full fidelity conversion", noteStyle)
	}
}
