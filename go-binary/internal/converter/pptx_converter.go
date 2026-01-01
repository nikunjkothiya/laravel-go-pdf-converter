package converter

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/nikunjkothiya/gopdfconv/internal/pdf"
	"github.com/nikunjkothiya/gopdfconv/pkg/errors"
)

// PPTXConverter handles PowerPoint (PPTX) to PDF conversion
type PPTXConverter struct {
	opts            pdf.Options
	libreOfficePath string
	useLibreOffice  bool
	forceNative     bool
}

// NewPPTXConverter creates a new PPTX converter
func NewPPTXConverter() *PPTXConverter {
	c := &PPTXConverter{
		opts: pdf.DefaultOptions(),
	}
	c.detectLibreOffice()
	return c
}

// SupportedExtensions returns extensions handled by this converter
func (c *PPTXConverter) SupportedExtensions() []string {
	return []string{".pptx", ".ppt", ".odp"}
}


// detectLibreOffice checks for LibreOffice installation
func (c *PPTXConverter) detectLibreOffice() {
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

// SetForceNative forces native conversion even if LibreOffice is available
func (c *PPTXConverter) SetForceNative(force bool) {
	c.forceNative = force
}

// Validate checks if the input file is a valid PPTX
func (c *PPTXConverter) Validate(inputPath string) error {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return errors.NewWithFile(errors.ErrFileNotFound, "File not found", inputPath)
	}

	r, err := zip.OpenReader(inputPath)
	if err != nil {
		return errors.NewWithDetails(errors.ErrInvalidFormat, "Not a valid PPTX file (invalid ZIP)", inputPath, err.Error())
	}
	defer r.Close()

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
	if err := c.Validate(inputPath); err != nil {
		return err
	}

	// Use LibreOffice if available and not forced to native
	if c.useLibreOffice && !c.forceNative {
		err := c.convertWithLibreOffice(inputPath, outputPath)
		if err == nil {
			return nil
		}
		// Fall back to native if LibreOffice fails
	}

	// Native Go conversion with improved rendering
	return c.convertNative(inputPath, outputPath, opts)
}

// convertWithLibreOffice uses LibreOffice for high-fidelity conversion
func (c *PPTXConverter) convertWithLibreOffice(inputPath, outputPath string) error {
	loConverter := NewLibreOfficeConverter(c.libreOfficePath)
	return loConverter.Convert(inputPath, outputPath)
}

// convertNative performs native Go conversion with improved slide rendering
func (c *PPTXConverter) convertNative(inputPath, outputPath string, opts pdf.Options) error {
	r, err := zip.OpenReader(inputPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to open PPTX")
	}
	defer r.Close()

	// Create temp directory for extracted images
	tempDir, err := os.MkdirTemp("", "pptx-images-*")
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create temp directory")
	}
	defer os.RemoveAll(tempDir)

	// Extract images from PPTX
	imageMap := c.extractImages(r, tempDir)

	// Parse slide relationships to map images
	relMap := c.parseRelationships(r)

	// Parse slides with full content
	slides, err := c.parseSlides(r, imageMap, relMap)
	if err != nil {
		return err
	}

	// Get slide dimensions from presentation.xml
	slideWidth, slideHeight := c.getSlideSize(r)

	// For PowerPoint, use only general options (page size, margins, watermark, header/footer)
	// Ignore table-specific customization options (they only apply to spreadsheets)
	pptOpts := c.sanitizeOptionsForPPT(opts)
	
	// Create PDF with landscape orientation for slides
	pptOpts.Orientation = pdf.Landscape
	builder, err := pdf.NewBuilder(pptOpts)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create PDF builder")
	}

	// Render each slide
	for i, slide := range slides {
		if i > 0 {
			builder.AddPage()
		} else {
			builder.AddPage()
		}

		c.renderSlideEnhanced(builder, slide, pptOpts, slideWidth, slideHeight, tempDir)
	}

	if err := builder.Save(outputPath); err != nil {
		return errors.Wrap(err, errors.ErrWriteFailed, "Failed to save PDF")
	}

	return nil
}


// Slide represents a parsed PowerPoint slide with full content
type Slide struct {
	Index      int
	Title      string
	Texts      []SlideText
	Images     []SlideImage
	Background SlideBackground
	Notes      string
}

// SlideText represents text on a slide with positioning
type SlideText struct {
	Content   string
	X, Y      float64 // Position in EMUs (English Metric Units)
	Width     float64
	Height    float64
	FontSize  float64
	Bold      bool
	Italic    bool
	Alignment string
	Color     string // Hex color like "FFFFFF"
	IsTitle   bool
}

// SlideImage represents an image on a slide
type SlideImage struct {
	RelID    string
	FilePath string
	X, Y     float64
	Width    float64
	Height   float64
}

// SlideBackground represents slide background
type SlideBackground struct {
	Color     string // Hex color
	ImagePath string
	HasColor  bool
	HasImage  bool
}

// extractImages extracts all images from PPTX to temp directory
func (c *PPTXConverter) extractImages(r *zip.ReadCloser, tempDir string) map[string]string {
	imageMap := make(map[string]string)

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/media/") {
			ext := strings.ToLower(filepath.Ext(f.Name))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".bmp" {
				rc, err := f.Open()
				if err != nil {
					continue
				}

				baseName := filepath.Base(f.Name)
				outPath := filepath.Join(tempDir, baseName)
				outFile, err := os.Create(outPath)
				if err != nil {
					rc.Close()
					continue
				}

				io.Copy(outFile, rc)
				outFile.Close()
				rc.Close()

				imageMap[baseName] = outPath
			}
		}
	}

	return imageMap
}

// parseRelationships parses slide relationships to map rId to image files
func (c *PPTXConverter) parseRelationships(r *zip.ReadCloser) map[string]map[string]string {
	relMap := make(map[string]map[string]string)

	for _, f := range r.File {
		if strings.Contains(f.Name, "_rels/slide") && strings.HasSuffix(f.Name, ".rels") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, _ := io.ReadAll(rc)
			rc.Close()

			// Extract slide number from path
			slideNum := ""
			re := regexp.MustCompile(`slide(\d+)\.xml\.rels`)
			if matches := re.FindStringSubmatch(f.Name); len(matches) > 1 {
				slideNum = matches[1]
			}

			if slideNum == "" {
				continue
			}

			relMap[slideNum] = make(map[string]string)

			// Parse relationships XML
			type Relationship struct {
				Id     string `xml:"Id,attr"`
				Target string `xml:"Target,attr"`
			}
			type Relationships struct {
				Rels []Relationship `xml:"Relationship"`
			}

			var rels Relationships
			xml.Unmarshal(data, &rels)

			for _, rel := range rels.Rels {
				if strings.Contains(rel.Target, "media/") {
					relMap[slideNum][rel.Id] = filepath.Base(rel.Target)
				}
			}
		}
	}

	return relMap
}


// getSlideSize extracts slide dimensions from presentation.xml
func (c *PPTXConverter) getSlideSize(r *zip.ReadCloser) (float64, float64) {
	// Default slide size (standard 16:9)
	defaultWidth := 9144000.0  // EMUs
	defaultHeight := 5143500.0 // EMUs

	for _, f := range r.File {
		if f.Name == "ppt/presentation.xml" {
			rc, err := f.Open()
			if err != nil {
				return defaultWidth, defaultHeight
			}
			data, _ := io.ReadAll(rc)
			rc.Close()

			// Extract sldSz (slide size)
			re := regexp.MustCompile(`<p:sldSz[^>]*cx="(\d+)"[^>]*cy="(\d+)"`)
			if matches := re.FindSubmatch(data); len(matches) > 2 {
				if w, err := strconv.ParseFloat(string(matches[1]), 64); err == nil {
					defaultWidth = w
				}
				if h, err := strconv.ParseFloat(string(matches[2]), 64); err == nil {
					defaultHeight = h
				}
			}
			break
		}
	}

	return defaultWidth, defaultHeight
}

// parseSlides extracts slide information from PPTX with full content
func (c *PPTXConverter) parseSlides(r *zip.ReadCloser, imageMap map[string]string, relMap map[string]map[string]string) ([]Slide, error) {
	var slides []Slide

	slideFiles := make(map[int]*zip.File)
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") && !strings.Contains(f.Name, "_rels") {
			numStr := strings.TrimPrefix(f.Name, "ppt/slides/slide")
			numStr = strings.TrimSuffix(numStr, ".xml")
			if num, err := strconv.Atoi(numStr); err == nil {
				slideFiles[num] = f
			}
		}
	}

	var slideNums []int
	for num := range slideFiles {
		slideNums = append(slideNums, num)
	}
	sort.Ints(slideNums)

	for _, num := range slideNums {
		slideFile := slideFiles[num]
		slide, err := c.parseSlideXMLEnhanced(slideFile, imageMap, relMap[strconv.Itoa(num)])
		if err != nil {
			continue
		}
		slide.Index = num
		slides = append(slides, slide)
	}

	return slides, nil
}

// PPTX XML structures for enhanced parsing
type slideXMLEnhanced struct {
	XMLName xml.Name `xml:"sld"`
	CSld    struct {
		Bg *struct {
			BgPr *struct {
				SolidFill *struct {
					SrgbClr *struct {
						Val string `xml:"val,attr"`
					} `xml:"srgbClr"`
				} `xml:"solidFill"`
			} `xml:"bgPr"`
		} `xml:"bg"`
		SpTree struct {
			Sp  []shapeXMLEnhanced `xml:"sp"`
			Pic []picXML           `xml:"pic"`
		} `xml:"spTree"`
	} `xml:"cSld"`
}

type shapeXMLEnhanced struct {
	NvSpPr struct {
		NvPr struct {
			Ph *struct {
				Type string `xml:"type,attr"`
			} `xml:"ph"`
		} `xml:"nvPr"`
	} `xml:"nvSpPr"`
	SpPr *struct {
		Xfrm *struct {
			Off *struct {
				X string `xml:"x,attr"`
				Y string `xml:"y,attr"`
			} `xml:"off"`
			Ext *struct {
				Cx string `xml:"cx,attr"`
				Cy string `xml:"cy,attr"`
			} `xml:"ext"`
		} `xml:"xfrm"`
	} `xml:"spPr"`
	TxBody *struct {
		P []paragraphXMLEnhanced `xml:"p"`
	} `xml:"txBody"`
}

type paragraphXMLEnhanced struct {
	PPr *struct {
		Algn string `xml:"algn,attr"`
	} `xml:"pPr"`
	R []runXMLEnhanced `xml:"r"`
}

type runXMLEnhanced struct {
	RPr *struct {
		Sz   string `xml:"sz,attr"`
		B    string `xml:"b,attr"`
		I    string `xml:"i,attr"`
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

type picXML struct {
	BlipFill *struct {
		Blip *struct {
			Embed string `xml:"embed,attr"`
		} `xml:"blip"`
	} `xml:"blipFill"`
	SpPr *struct {
		Xfrm *struct {
			Off *struct {
				X string `xml:"x,attr"`
				Y string `xml:"y,attr"`
			} `xml:"off"`
			Ext *struct {
				Cx string `xml:"cx,attr"`
				Cy string `xml:"cy,attr"`
			} `xml:"ext"`
		} `xml:"xfrm"`
	} `xml:"spPr"`
}


// parseSlideXMLEnhanced parses a single slide XML file with full content
func (c *PPTXConverter) parseSlideXMLEnhanced(f *zip.File, imageMap map[string]string, slideRels map[string]string) (Slide, error) {
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

	var sld slideXMLEnhanced
	if err := xml.Unmarshal(data, &sld); err != nil {
		// Fall back to simple text extraction
		return c.extractTextSimple(data), nil
	}

	// Extract background color
	if sld.CSld.Bg != nil && sld.CSld.Bg.BgPr != nil && sld.CSld.Bg.BgPr.SolidFill != nil {
		if sld.CSld.Bg.BgPr.SolidFill.SrgbClr != nil {
			slide.Background.Color = sld.CSld.Bg.BgPr.SolidFill.SrgbClr.Val
			slide.Background.HasColor = true
		}
	}

	// Extract text from shapes
	for _, sp := range sld.CSld.SpTree.Sp {
		if sp.TxBody == nil {
			continue
		}

		var textContent strings.Builder
		var fontSize float64 = 12
		var color string
		var bold, italic bool
		var alignment string

		for _, p := range sp.TxBody.P {
			if p.PPr != nil && p.PPr.Algn != "" {
				alignment = p.PPr.Algn
			}
			for _, r := range p.R {
				textContent.WriteString(r.T)
				if r.RPr != nil {
					if r.RPr.Sz != "" {
						if sz, err := strconv.ParseFloat(r.RPr.Sz, 64); err == nil {
							fontSize = sz / 100 // Convert from hundredths of a point
						}
					}
					if r.RPr.B == "1" {
						bold = true
					}
					if r.RPr.I == "1" {
						italic = true
					}
					if r.RPr.SolidFill != nil {
						if r.RPr.SolidFill.SrgbClr != nil {
							color = r.RPr.SolidFill.SrgbClr.Val
						} else if r.RPr.SolidFill.SchemeClr != nil {
							color = c.mapSchemeColor(r.RPr.SolidFill.SchemeClr.Val)
						}
					}
				}
			}
			textContent.WriteString("\n")
		}

		text := strings.TrimSpace(textContent.String())
		if text == "" {
			continue
		}

		// Get position
		var x, y, w, h float64
		if sp.SpPr != nil && sp.SpPr.Xfrm != nil {
			if sp.SpPr.Xfrm.Off != nil {
				x, _ = strconv.ParseFloat(sp.SpPr.Xfrm.Off.X, 64)
				y, _ = strconv.ParseFloat(sp.SpPr.Xfrm.Off.Y, 64)
			}
			if sp.SpPr.Xfrm.Ext != nil {
				w, _ = strconv.ParseFloat(sp.SpPr.Xfrm.Ext.Cx, 64)
				h, _ = strconv.ParseFloat(sp.SpPr.Xfrm.Ext.Cy, 64)
			}
		}

		isTitle := sp.NvSpPr.NvPr.Ph != nil &&
			(sp.NvSpPr.NvPr.Ph.Type == "title" || sp.NvSpPr.NvPr.Ph.Type == "ctrTitle")

		if isTitle && slide.Title == "" {
			slide.Title = text
		}

		slide.Texts = append(slide.Texts, SlideText{
			Content:   text,
			X:         x,
			Y:         y,
			Width:     w,
			Height:    h,
			FontSize:  fontSize,
			Bold:      bold,
			Italic:    italic,
			Alignment: alignment,
			Color:     color,
			IsTitle:   isTitle,
		})
	}

	// Extract images
	for _, pic := range sld.CSld.SpTree.Pic {
		if pic.BlipFill == nil || pic.BlipFill.Blip == nil {
			continue
		}

		relId := pic.BlipFill.Blip.Embed
		if relId == "" {
			continue
		}

		// Look up image file from relationships
		imageName := slideRels[relId]
		if imageName == "" {
			continue
		}

		imagePath := imageMap[imageName]
		if imagePath == "" {
			continue
		}

		var x, y, w, h float64
		if pic.SpPr != nil && pic.SpPr.Xfrm != nil {
			if pic.SpPr.Xfrm.Off != nil {
				x, _ = strconv.ParseFloat(pic.SpPr.Xfrm.Off.X, 64)
				y, _ = strconv.ParseFloat(pic.SpPr.Xfrm.Off.Y, 64)
			}
			if pic.SpPr.Xfrm.Ext != nil {
				w, _ = strconv.ParseFloat(pic.SpPr.Xfrm.Ext.Cx, 64)
				h, _ = strconv.ParseFloat(pic.SpPr.Xfrm.Ext.Cy, 64)
			}
		}

		slide.Images = append(slide.Images, SlideImage{
			RelID:    relId,
			FilePath: imagePath,
			X:        x,
			Y:        y,
			Width:    w,
			Height:   h,
		})
	}

	return slide, nil
}

// mapSchemeColor maps PowerPoint scheme colors to hex values
func (c *PPTXConverter) mapSchemeColor(scheme string) string {
	colorMap := map[string]string{
		"bg1":     "FFFFFF",
		"bg2":     "E7E6E6",
		"tx1":     "000000",
		"tx2":     "44546A",
		"accent1": "4472C4",
		"accent2": "ED7D31",
		"accent3": "A5A5A5",
		"accent4": "FFC000",
		"accent5": "5B9BD5",
		"accent6": "70AD47",
		"lt1":     "FFFFFF",
		"lt2":     "E7E6E6",
		"dk1":     "000000",
		"dk2":     "44546A",
	}
	if color, ok := colorMap[scheme]; ok {
		return color
	}
	return "000000"
}

// extractTextSimple uses regex for simple text extraction as fallback
func (c *PPTXConverter) extractTextSimple(data []byte) Slide {
	slide := Slide{}

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


// renderSlideEnhanced renders a slide to PDF with images and better layout
func (c *PPTXConverter) renderSlideEnhanced(builder *pdf.Builder, slide Slide, opts pdf.Options, slideW, slideH float64, tempDir string) {
	// Calculate scale factor from EMUs to PDF points
	// EMUs: 914400 per inch, PDF points: 72 per inch
	pageWidth := opts.PageSize.Height  // Landscape
	pageHeight := opts.PageSize.Width
	if opts.Orientation == pdf.Portrait {
		pageWidth = opts.PageSize.Width
		pageHeight = opts.PageSize.Height
	}

	contentWidth := pageWidth - (opts.Margin * 2)
	contentHeight := pageHeight - (opts.Margin * 2) - 40 // Leave room for header/footer

	scaleX := contentWidth / slideW
	scaleY := contentHeight / slideH
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// EMU to points conversion (914400 EMUs = 1 inch = 72 points)
	emuToPoints := func(emu float64) float64 {
		return (emu / 914400.0) * 72.0 * (scale * 914400.0 / 72.0)
	}

	// Draw background color if present
	if slide.Background.HasColor && slide.Background.Color != "" {
		bgColor := pdf.ParseHexColor(slide.Background.Color)
		builder.SetFillColor(bgColor)
		pdfObj := builder.GetPdf()
		pdfObj.Rectangle(opts.Margin, opts.Margin+20, pageWidth-opts.Margin, pageHeight-opts.Margin-10, "F", 0, 0)
	}

	// Draw images first (background layer)
	for _, img := range slide.Images {
		if img.FilePath == "" {
			continue
		}

		// Check if image file exists and is valid
		if _, err := os.Stat(img.FilePath); err != nil {
			continue
		}

		// Calculate position and size
		imgX := opts.Margin + emuToPoints(img.X)
		imgY := opts.Margin + 20 + emuToPoints(img.Y)
		imgW := emuToPoints(img.Width)
		imgH := emuToPoints(img.Height)

		// Ensure reasonable bounds
		if imgW < 10 {
			imgW = 100
		}
		if imgH < 10 {
			imgH = 100
		}
		if imgW > contentWidth {
			imgW = contentWidth
		}
		if imgH > contentHeight {
			imgH = contentHeight
		}

		// Try to get actual image dimensions for aspect ratio
		if file, err := os.Open(img.FilePath); err == nil {
			if imgConfig, _, err := image.DecodeConfig(file); err == nil {
				aspectRatio := float64(imgConfig.Width) / float64(imgConfig.Height)
				if imgW/imgH > aspectRatio {
					imgW = imgH * aspectRatio
				} else {
					imgH = imgW / aspectRatio
				}
			}
			file.Close()
		}

		builder.AddImage(img.FilePath, imgX, imgY, imgW, imgH)
	}

	// Sort texts by Y position (top to bottom)
	sortedTexts := make([]SlideText, len(slide.Texts))
	copy(sortedTexts, slide.Texts)
	sort.Slice(sortedTexts, func(i, j int) bool {
		return sortedTexts[i].Y < sortedTexts[j].Y
	})

	// Draw text elements
	for _, text := range sortedTexts {
		if text.Content == "" {
			continue
		}

		// Calculate position
		textX := opts.Margin + emuToPoints(text.X)
		textY := opts.Margin + 20 + emuToPoints(text.Y)

		// Ensure text is within bounds
		if textX < opts.Margin {
			textX = opts.Margin
		}
		if textY < opts.Margin+20 {
			textY = opts.Margin + 20
		}

		// Determine font size
		fontSize := text.FontSize
		if fontSize < 8 {
			fontSize = 12
		}
		if fontSize > 48 {
			fontSize = 48
		}

		// Title gets larger font
		if text.IsTitle {
			fontSize = 24
			if text.FontSize > 24 {
				fontSize = text.FontSize
			}
		}

		// Create text style
		style := pdf.DefaultStyle()
		style.FontSize = fontSize
		if text.Bold {
			style.FontStyle = "B"
		}

		// Handle text color with smart fallback
		if text.Color != "" {
			style.TextColor = pdf.ParseHexColor(text.Color)
			// Smart color fallback: if text is white/very light, make it dark
			if style.TextColor.R > 240 && style.TextColor.G > 240 && style.TextColor.B > 240 {
				style.TextColor = pdf.ColorBlack
			}
		}

		builder.SetXY(textX, textY)
		builder.SetFont(style.FontFamily, style.FontStyle, style.FontSize)
		builder.SetTextColor(style.TextColor)

		// Draw text lines
		lines := strings.Split(text.Content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			builder.GetPdf().Text(line)
			textY += fontSize + 4
			builder.SetXY(textX, textY)
		}
	}

	// Add slide number
	slideNumStyle := pdf.DefaultStyle()
	slideNumStyle.FontSize = 10
	slideNumStyle.TextColor = pdf.ColorGray
	builder.SetXY(pageWidth-opts.Margin-40, pageHeight-opts.Margin)
	builder.SetFont(slideNumStyle.FontFamily, "", slideNumStyle.FontSize)
	builder.SetTextColor(slideNumStyle.TextColor)
	builder.GetPdf().Text(fmt.Sprintf("Slide %d", slide.Index))
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

// sanitizeOptionsForPPT returns options with only general settings applied
// Table-specific customization options (styling, row/cell settings) are reset to defaults
// as they only apply to spreadsheet formats (CSV, XLS, XLSX)
func (c *PPTXConverter) sanitizeOptionsForPPT(opts pdf.Options) pdf.Options {
	// Start with default options
	pptOpts := pdf.DefaultOptions()
	
	// Keep general page options
	pptOpts.PageSize = opts.PageSize
	pptOpts.Orientation = opts.Orientation
	pptOpts.Margin = opts.Margin
	pptOpts.FontFamily = opts.FontFamily
	pptOpts.FontSize = opts.FontSize
	
	// Keep metadata options
	pptOpts.Title = opts.Title
	pptOpts.Author = opts.Author
	pptOpts.Subject = opts.Subject
	
	// Keep header/footer options
	pptOpts.HeaderText = opts.HeaderText
	pptOpts.FooterText = opts.FooterText
	
	// Keep watermark options
	pptOpts.CustomFontPath = opts.CustomFontPath
	pptOpts.WatermarkText = opts.WatermarkText
	pptOpts.WatermarkImage = opts.WatermarkImage
	pptOpts.WatermarkAlpha = opts.WatermarkAlpha
	
	// Keep quality options
	pptOpts.Compression = opts.Compression
	pptOpts.Quality = opts.Quality
	
	// Ignore table-specific options (use defaults):
	// - HeaderColor, HeaderTextColor, RowColor, RowTextColor, BorderColor
	// - ShowGridLines, RowHeight, HeaderHeight, CellPadding
	// - MinColumnWidth, MaxColumnWidth, HeaderFontSize, HeaderFontBold
	// - HeaderRow, AutoWidth, AutoOrientation
	
	return pptOpts
}
