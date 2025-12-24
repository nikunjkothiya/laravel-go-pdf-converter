package converter

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/nikunjkothiya/gopdfconv/internal/pdf"
	"github.com/nikunjkothiya/gopdfconv/pkg/errors"
)

// PPTXConverter handles PowerPoint (PPTX) to PDF conversion
type PPTXConverter struct {
	opts             pdf.Options
	libreOfficePath  string
	useLibreOffice   bool
}

// NewPPTXConverter creates a new PPTX converter
func NewPPTXConverter() *PPTXConverter {
	c := &PPTXConverter{
		opts: pdf.DefaultOptions(),
	}
	// Try to detect LibreOffice installation
	c.detectLibreOffice()
	return c
}

// SupportedExtensions returns extensions handled by this converter
func (c *PPTXConverter) SupportedExtensions() []string {
	return []string{".pptx", ".ppt", ".odp"}
}

// detectLibreOffice checks for LibreOffice installation
func (c *PPTXConverter) detectLibreOffice() {
	// Common LibreOffice paths
	paths := []string{
		"/usr/bin/libreoffice",
		"/usr/bin/soffice",
		"/usr/local/bin/libreoffice",
		"/Applications/LibreOffice.app/Contents/MacOS/soffice",
		"C:\\Program Files\\LibreOffice\\program\\soffice.exe",
		"C:\\Program Files (x86)\\LibreOffice\\program\\soffice.exe",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			c.libreOfficePath = p
			c.useLibreOffice = true
			return
		}
	}

	// Try to find in PATH
	if path, err := exec.LookPath("libreoffice"); err == nil {
		c.libreOfficePath = path
		c.useLibreOffice = true
		return
	}
	if path, err := exec.LookPath("soffice"); err == nil {
		c.libreOfficePath = path
		c.useLibreOffice = true
		return
	}
}

// SetLibreOfficePath manually sets the LibreOffice path
func (c *PPTXConverter) SetLibreOfficePath(path string) {
	if _, err := os.Stat(path); err == nil {
		c.libreOfficePath = path
		c.useLibreOffice = true
	}
}

// Validate checks if the input file is a valid PPTX
func (c *PPTXConverter) Validate(inputPath string) error {
	// Check file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return errors.NewWithFile(errors.ErrFileNotFound, "File not found", inputPath)
	}

	// Check if it's a valid ZIP (PPTX is a ZIP file)
	r, err := zip.OpenReader(inputPath)
	if err != nil {
		return errors.NewWithDetails(errors.ErrInvalidFormat, "Not a valid PPTX file (invalid ZIP)", inputPath, err.Error())
	}
	defer r.Close()

	// Check for required PPTX structure
	hasContentTypes := false
	hasPresentationXML := false

	for _, f := range r.File {
		if f.Name == "[Content_Types].xml" {
			hasContentTypes = true
		}
		if strings.Contains(f.Name, "ppt/presentation.xml") {
			hasPresentationXML = true
		}
	}

	if !hasContentTypes || !hasPresentationXML {
		return errors.NewWithFile(errors.ErrInvalidFormat, "Not a valid PPTX file (missing required components)", inputPath)
	}

	return nil
}

// Convert performs the PPTX to PDF conversion
func (c *PPTXConverter) Convert(inputPath, outputPath string, opts pdf.Options) error {
	// Validate input
	if err := c.Validate(inputPath); err != nil {
		return err
	}

	// Use LibreOffice if available (best fidelity)
	if c.useLibreOffice {
		return c.convertWithLibreOffice(inputPath, outputPath)
	}

	// Fall back to native Go conversion (limited fidelity)
	return c.convertNative(inputPath, outputPath, opts)
}

// convertWithLibreOffice uses LibreOffice for high-fidelity conversion
func (c *PPTXConverter) convertWithLibreOffice(inputPath, outputPath string) error {
	loConverter := NewLibreOfficeConverter(c.libreOfficePath)
	return loConverter.Convert(inputPath, outputPath)
}

// copyFile copies a file from src to dst
func (c *PPTXConverter) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// convertNative performs native Go conversion (limited fidelity)
func (c *PPTXConverter) convertNative(inputPath, outputPath string, opts pdf.Options) error {
	// Open PPTX as ZIP
	r, err := zip.OpenReader(inputPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to open PPTX")
	}
	defer r.Close()

	// Parse slides
	slides, err := c.parseSlides(r)
	if err != nil {
		return err
	}

	// Create PDF
	builder, err := pdf.NewBuilder(opts)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create PDF builder")
	}

	// Render each slide as a page
	for i, slide := range slides {
		if i > 0 {
			builder.AddPage()
		} else {
			builder.AddPage()
		}

		// Render slide content
		c.renderSlide(builder, slide, opts)
	}

	// Save PDF
	if err := builder.Save(outputPath); err != nil {
		return errors.Wrap(err, errors.ErrWriteFailed, "Failed to save PDF")
	}

	return nil
}

// Slide represents a parsed PowerPoint slide
type Slide struct {
	Index    int
	Title    string
	Texts    []SlideText
	Images   []SlideImage
	Notes    string
}

// SlideText represents text on a slide
type SlideText struct {
	Content   string
	X, Y      float64 // Position as percentage of slide
	Width     float64
	FontSize  float64
	Bold      bool
	Italic    bool
	Alignment string
	Color     string // Hex color like "FFFFFF"
}

// SlideImage represents an image on a slide
type SlideImage struct {
	RelID   string
	X, Y    float64
	Width   float64
	Height  float64
}

// parseSlides extracts slide information from PPTX
func (c *PPTXConverter) parseSlides(r *zip.ReadCloser) ([]Slide, error) {
	var slides []Slide

	// Find all slide XML files
	slideFiles := make(map[int]*zip.File)
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			// Extract slide number
			numStr := strings.TrimPrefix(f.Name, "ppt/slides/slide")
			numStr = strings.TrimSuffix(numStr, ".xml")
			if num, err := strconv.Atoi(numStr); err == nil {
				slideFiles[num] = f
			}
		}
	}

	// Sort slides by number
	var slideNums []int
	for num := range slideFiles {
		slideNums = append(slideNums, num)
	}
	sort.Ints(slideNums)

	// Parse each slide
	for _, num := range slideNums {
		slideFile := slideFiles[num]
		slide, err := c.parseSlideXML(slideFile)
		if err != nil {
			continue // Skip slides that fail to parse
		}
		slide.Index = num
		slides = append(slides, slide)
	}

	return slides, nil
}

// PPTX XML structures for parsing
type slideXML struct {
	CSld struct {
		SpTree struct {
			Sp []shapeXML `xml:"sp"`
		} `xml:"spTree"`
	} `xml:"cSld"`
}

type shapeXML struct {
	NvSpPr struct {
		NvPr struct {
			Ph *struct {
				Type string `xml:"type,attr"`
			} `xml:"ph"`
		} `xml:"nvPr"`
	} `xml:"nvSpPr"`
	TxBody *struct {
		P []paragraphXML `xml:"p"`
	} `xml:"txBody"`
}

type paragraphXML struct {
	R []runXML `xml:"r"`
}

type runXML struct {
	RPr *struct {
		SolidFill *struct {
			SrgbClr *struct {
				Val string `xml:"val,attr"`
			} `xml:"srgbClr"`
			SchemeClr *struct {
				Val string `xml:"val,attr"`
			} `xml:"schemeClr"`
		} `xml:"solidFill"`
	} `xml:"rPr"`
	T string `xml:"t"`
}

// parseSlideXML parses a single slide XML file
func (c *PPTXConverter) parseSlideXML(f *zip.File) (Slide, error) {
	slide := Slide{}

	rc, err := f.Open()
	if err != nil {
		return slide, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return slide, err
	}

	var sld slideXML
	if err := xml.Unmarshal(data, &sld); err != nil {
		// Try simpler text extraction
		return c.extractTextSimple(data), nil
	}

	// Extract text from shapes
	for _, sp := range sld.CSld.SpTree.Sp {
		if sp.TxBody == nil {
			continue
		}

		var textContent strings.Builder
		for _, p := range sp.TxBody.P {
			for _, r := range p.R {
				textContent.WriteString(r.T)
			}
			textContent.WriteString("\n")
		}

		text := strings.TrimSpace(textContent.String())
		if text != "" {
			// Extract color from first run
			color := ""
			if len(sp.TxBody.P) > 0 && len(sp.TxBody.P[0].R) > 0 {
				r := sp.TxBody.P[0].R[0]
				if r.RPr != nil && r.RPr.SolidFill != nil {
					if r.RPr.SolidFill.SrgbClr != nil {
						color = r.RPr.SolidFill.SrgbClr.Val
					} else if r.RPr.SolidFill.SchemeClr != nil {
						// Map scheme colors (simplified)
						if r.RPr.SolidFill.SchemeClr.Val == "bg1" {
							color = "FFFFFF"
						} else if r.RPr.SolidFill.SchemeClr.Val == "tx1" {
							color = "000000"
						}
					}
				}
			}

			// Check if this is a title
			isTitle := sp.NvSpPr.NvPr.Ph != nil && 
				(sp.NvSpPr.NvPr.Ph.Type == "title" || sp.NvSpPr.NvPr.Ph.Type == "ctrTitle")

			if isTitle && slide.Title == "" {
				slide.Title = text
			} else {
				slide.Texts = append(slide.Texts, SlideText{
					Content:  text,
					FontSize: 12,
					Color:    color,
				})
			}
		}
	}

	return slide, nil
}

// extractTextSimple uses regex for simple text extraction as fallback
func (c *PPTXConverter) extractTextSimple(data []byte) Slide {
	slide := Slide{}

	// Simple regex to extract text content
	re := regexp.MustCompile(`<a:t>([^<]+)</a:t>`)
	matches := re.FindAllSubmatch(data, -1)

	var texts []string
	for _, match := range matches {
		if len(match) > 1 {
			text := strings.TrimSpace(string(match[1]))
			if text != "" {
				texts = append(texts, text)
			}
		}
	}

	if len(texts) > 0 {
		slide.Title = texts[0]
		for i := 1; i < len(texts); i++ {
			slide.Texts = append(slide.Texts, SlideText{
				Content:  texts[i],
				FontSize: 12,
			})
		}
	}

	return slide
}

// renderSlide renders a slide to PDF
func (c *PPTXConverter) renderSlide(builder *pdf.Builder, slide Slide, opts pdf.Options) {
	style := pdf.DefaultStyle()
	titleStyle := pdf.HeaderStyle()
	titleStyle.FontSize = 24

	// Render title
	if slide.Title != "" {
		builder.AddText(slide.Title, titleStyle)
		builder.NewLine(20)
	}

	// Render other text elements
	for _, text := range slide.Texts {
		textStyle := style
		if text.Bold {
			textStyle.FontStyle = "B"
		}
		if text.FontSize > 0 {
			textStyle.FontSize = text.FontSize
		}

		// Handle color
		if text.Color != "" {
			textStyle.TextColor = pdf.ParseHexColor(text.Color)
			
			// SMART COLOR FALLBACK:
			// If text is white (or very light) and we are rendering on white background,
			// force it to black/dark gray so it's visible.
			if textStyle.TextColor.R > 240 && textStyle.TextColor.G > 240 && textStyle.TextColor.B > 240 {
				textStyle.TextColor = pdf.ColorBlack
			}
		}

		builder.AddText(text.Content, textStyle)
		builder.NewLine(textStyle.FontSize + 4)
	}

	// Add slide number
	slideNumStyle := pdf.DefaultStyle()
	slideNumStyle.FontSize = 8
	slideNumStyle.TextColor = pdf.ColorGray
	builder.SetXY(opts.ContentWidth()-20, opts.ContentHeight()-10)
	builder.AddText(fmt.Sprintf("Slide %d", slide.Index), slideNumStyle)
}

// HasLibreOffice returns whether LibreOffice is available
func (c *PPTXConverter) HasLibreOffice() bool {
	return c.useLibreOffice
}

// GetLibreOfficePath returns the detected LibreOffice path
func (c *PPTXConverter) GetLibreOfficePath() string {
	return c.libreOfficePath
}

// SetUseLibreOffice enables or disables LibreOffice usage
func (c *PPTXConverter) SetUseLibreOffice(use bool) {
	c.useLibreOffice = use
}
