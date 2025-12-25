package converter

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strings"

	"github.com/nikunjkothiya/gopdfconv/internal/pdf"
	"github.com/nikunjkothiya/gopdfconv/pkg/errors"
)

// CSVConverter handles CSV to PDF conversion with streaming support
type CSVConverter struct {
	opts          pdf.Options
	maxSampleRows int // Number of rows to sample for column width calculation
	onProgress    func(int)
}

// NewCSVConverter creates a new CSV converter
func NewCSVConverter() *CSVConverter {
	return &CSVConverter{
		opts:          pdf.DefaultOptions(),
		maxSampleRows: 100,
	}
}

// SetProgressCallback sets the callback for progress reporting
func (c *CSVConverter) SetProgressCallback(callback func(int)) {
	c.onProgress = callback
}

// SupportedExtensions returns extensions handled by this converter
func (c *CSVConverter) SupportedExtensions() []string {
	return []string{".csv", ".tsv", ".txt"}
}

// Validate checks if the input file is a valid CSV
func (c *CSVConverter) Validate(inputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return errors.NewWithFile(errors.ErrFileNotFound, "Cannot open file", inputPath)
	}
	defer file.Close()

	// Check if file is readable as CSV
	reader := csv.NewReader(bufio.NewReader(file))
	reader.FieldsPerRecord = -1 // Allow variable field counts
	reader.LazyQuotes = true    // Be lenient with quotes

	// Try to read first few rows
	for i := 0; i < 5; i++ {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.NewWithDetails(errors.ErrInvalidFormat, "Invalid CSV format", inputPath, err.Error())
		}
	}

	return nil
}

// Convert performs the CSV to PDF conversion
func (c *CSVConverter) Convert(inputPath, outputPath string, opts pdf.Options) error {
	// Validate input
	if err := c.Validate(inputPath); err != nil {
		return err
	}

	// Open file for reading
	file, err := os.Open(inputPath)
	if err != nil {
		return errors.NewWithFile(errors.ErrFileNotFound, "Cannot open input file", inputPath)
	}
	defer file.Close()

	// Create buffered reader for efficient streaming
	bufferedReader := bufio.NewReaderSize(file, 64*1024) // 64KB buffer
	
	// Skip UTF-8 BOM if present
	bom := make([]byte, 3)
	n, err := bufferedReader.Read(bom)
	if err != nil && err != io.EOF {
		return errors.NewWithFile(errors.ErrConversionFailed, "Failed to read file", inputPath)
	}
	// Check for UTF-8 BOM (0xEF, 0xBB, 0xBF)
	if n < 3 || bom[0] != 0xEF || bom[1] != 0xBB || bom[2] != 0xBF {
		// No BOM, reset the reader by seeking to start
		file.Seek(0, 0)
		bufferedReader = bufio.NewReaderSize(file, 64*1024)
	}
	
	reader := csv.NewReader(bufferedReader)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Detect delimiter (comma, tab, semicolon)
	reader.Comma = c.detectDelimiter(inputPath)

	// Read all records (for now, will optimize for streaming later)
	var allRecords [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip malformed rows but continue
			continue
		}
		allRecords = append(allRecords, record)
	}

	if len(allRecords) == 0 {
		return errors.NewWithFile(errors.ErrInvalidFormat, "CSV file is empty", inputPath)
	}

	// Calculate optimal column widths
	colWidths, shouldSwitchToLandscape := c.calculateColumnWidths(allRecords, opts)
	
	// Apply auto-orientation if needed
	if shouldSwitchToLandscape {
		opts.Orientation = pdf.Landscape
		// Re-optimize widths for new orientation
		colWidths, _ = c.calculateColumnWidths(allRecords, opts)
	}

	// Create PDF builder
	builder, err := pdf.NewBuilder(opts)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create PDF builder")
	}
	
	if c.onProgress != nil {
		builder.SetProgressCallback(c.onProgress)
	}

	// Add first page
	builder.AddPage()

	// Separate headers and data
	var headers []string
	var dataRows [][]string

	if opts.HeaderRow && len(allRecords) > 0 {
		headers = allRecords[0]
		if len(allRecords) > 1 {
			dataRows = allRecords[1:]
		}
	} else {
		dataRows = allRecords
	}

	// Draw the table
	if err := builder.DrawTable(headers, dataRows, colWidths); err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to draw table")
	}

	// Save the PDF
	if err := builder.Save(outputPath); err != nil {
		return errors.Wrap(err, errors.ErrWriteFailed, "Failed to save PDF")
	}

	return nil
}

// detectDelimiter attempts to detect the CSV delimiter
func (c *CSVConverter) detectDelimiter(filePath string) rune {
	file, err := os.Open(filePath)
	if err != nil {
		return ','
	}
	defer file.Close()

	// Read first line
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()

		// Count occurrences of common delimiters
		commaCount := strings.Count(line, ",")
		tabCount := strings.Count(line, "\t")
		semicolonCount := strings.Count(line, ";")
		pipeCount := strings.Count(line, "|")

		// Return the most common delimiter
		maxCount := commaCount
		delimiter := ','

		if tabCount > maxCount {
			maxCount = tabCount
			delimiter = '\t'
		}
		if semicolonCount > maxCount {
			maxCount = semicolonCount
			delimiter = ';'
		}
		if pipeCount > maxCount {
			delimiter = '|'
		}

		return delimiter
	}

	return ','
}

// calculateColumnWidths calculates optimal column widths based on content
// Returns widths and a boolean indicating if orientation should switch to Landscape
func (c *CSVConverter) calculateColumnWidths(records [][]string, opts pdf.Options) ([]float64, bool) {
	if len(records) == 0 {
		return nil, false
	}

	// Find the maximum number of columns
	maxCols := 0
	for _, row := range records {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	if maxCols == 0 {
		return nil, false
	}

	// Calculate max width for each column using accurate font measurement
	colMaxWidths := make([]float64, maxCols)
	sampleSize := c.maxSampleRows
	if len(records) < sampleSize {
		sampleSize = len(records)
	}

	// Create a temporary builder to access font measurement
	builder, err := pdf.NewBuilder(opts)
	if err != nil {
		// Fallback to estimation if font loading fails
		// We ignore auto-orientation in fallback for simplicity
		return c.calculateColumnWidthsFallback(records, opts), false
	}

	for i := 0; i < sampleSize; i++ {
		row := records[i]
		for j, cell := range row {
			// Accurate measurement + padding (left+right)
			width := builder.MeasureTextWidth(cell) + 6.0 // 3.0 padding per side
			if width > colMaxWidths[j] {
				colMaxWidths[j] = width
			}
		}
	}

	// Apply minimum width constraints
	const minColWidth = 40.0
	const maxColWidth = 250.0 // Increased max width for better readability

	for i := range colMaxWidths {
		if colMaxWidths[i] < minColWidth {
			colMaxWidths[i] = minColWidth
		}
		// Soft cap: allow going over if page permits, but clamp for initial calculation
	}
	
	// Check if we should switch to Landscape
	shouldSwitch := false
	if opts.AutoOrientation && opts.Orientation == pdf.Portrait {
		totalWidth := 0.0
		for _, w := range colMaxWidths {
			totalWidth += w
		}
		
		// If content is wider than Portrait content width but fits in Landscape, switch
		portraitWidth := opts.PageSize.Width - (opts.Margin * 2) // Current is portrait
		landscapeWidth := opts.PageSize.Height - (opts.Margin * 2)
		
		if totalWidth > portraitWidth && totalWidth <= landscapeWidth * 1.2 { 
			// Allow 20% overflow for landscape candidate to account for compression success
			shouldSwitch = true
		}
	}

	return c.optimizeWidthsForPage(colMaxWidths, opts.ContentWidth()), shouldSwitch
}

// optimizeWidthsForPage fits column widths to the page using weighted compression
func (c *CSVConverter) optimizeWidthsForPage(widths []float64, availableWidth float64) []float64 {
	totalWidth := 0.0
	for _, w := range widths {
		totalWidth += w
	}

	if totalWidth <= availableWidth {
		return widths
	}

	// Calculate how much we need to shave off
	overflow := totalWidth - availableWidth
	
	// Identify "wide" candidates (e.g. > 80pt) to preserve small data columns
	// Date columns are usually around 60-80pt, so anything above 80 is fair game using 80 as threshold
	threshold := 80.0 
	
	var wideIndices []int
	totalWideWidth := 0.0
	
	newWidths := make([]float64, len(widths))
	copy(newWidths, widths)
	
	for i, w := range widths {
		if w > threshold {
			wideIndices = append(wideIndices, i)
			totalWideWidth += w
		}
	}
	
	// Strategy 1: Reduce only wide columns proportionally
	if len(wideIndices) > 0 && totalWideWidth > (overflow + (float64(len(wideIndices)) * threshold)) {
		// We have enough "excess fat" in wide columns to absorb the overflow
		// while keeping them above threshold.
		
		// Calculate reduction ratio for wide columns
		// We want: (TotalWide - Reduction) + TotalNarrow = Available
		// Reduction = Overflow
		
		// Distribute overflow weighted by size
		for _, i := range wideIndices {
			// Weight = my_width / total_wide_width
			// share = overflow * weight
			weight := widths[i] / totalWideWidth
			share := overflow * weight
			newWidths[i] -= share
		}
		return newWidths
	}
	
	// Strategy 2: If Strategy 1 fails (overflow is huge), fall back to global scaling
	// but try to respect a hard floor for data columns.
	
	scale := availableWidth / totalWidth
	
	for i := range newWidths {
		newWidths[i] *= scale
		if newWidths[i] < 35 {
			newWidths[i] = 35
		}
	}
	
	// Check if floors pushed us over again
	newTotal := 0.0
	for _, w := range newWidths {
		newTotal += w
	}
	
	if newTotal > availableWidth {
		// Brutal truncate last resort
		finalScale := availableWidth / newTotal
		for i := range newWidths {
			newWidths[i] *= finalScale
		}
	}
	
	return newWidths
}

func (c *CSVConverter) calculateColumnWidthsFallback(records [][]string, opts pdf.Options) []float64 {
	if len(records) == 0 {
		return nil
	}
	
	// Basic character counting estimation
	maxCols := 0
	for _, row := range records {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	
	colMaxWidths := make([]float64, maxCols)
	sampleSize := c.maxSampleRows
	if len(records) < sampleSize {
		sampleSize = len(records)
	}
	
	for i := 0; i < sampleSize; i++ {
		row := records[i]
		for j, cell := range row {
			width := float64(len(cell)) * 6 // Estimate
			if width > colMaxWidths[j] {
				colMaxWidths[j] = width
			}
		}
	}
	
	// Apply standard scaling
	return c.optimizeWidthsForPage(colMaxWidths, opts.ContentWidth())
}

// StreamingCSVConverter provides memory-efficient conversion for large files
type StreamingCSVConverter struct {
	*CSVConverter
	chunkSize int
}

// NewStreamingCSVConverter creates a converter for large files
func NewStreamingCSVConverter(chunkSize int) *StreamingCSVConverter {
	return &StreamingCSVConverter{
		CSVConverter: NewCSVConverter(),
		chunkSize:    chunkSize,
	}
}

// ConvertStreaming performs memory-efficient streaming conversion
func (c *StreamingCSVConverter) ConvertStreaming(inputPath, outputPath string, opts pdf.Options) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return errors.NewWithFile(errors.ErrFileNotFound, "Cannot open input file", inputPath)
	}
	defer file.Close()

	// Get file size for progress tracking
	stat, _ := file.Stat()
	_ = stat.Size() // For future progress reporting

	// Create buffered reader
	reader := csv.NewReader(bufio.NewReaderSize(file, 64*1024))
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.Comma = c.detectDelimiter(inputPath)

	// First pass: sample rows for column width calculation
	var sampleRows [][]string
	for i := 0; i < c.maxSampleRows; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		sampleRows = append(sampleRows, record)
	}

	if len(sampleRows) == 0 {
		return errors.NewWithFile(errors.ErrInvalidFormat, "CSV file is empty", inputPath)
	}

	colWidths, shouldSwitchToLandscape := c.calculateColumnWidths(sampleRows, opts)
	
	if shouldSwitchToLandscape {
		opts.Orientation = pdf.Landscape
		// Re-calc for landscape
		colWidths, _ = c.calculateColumnWidths(sampleRows, opts)
	}

	// Reset file for second pass
	file.Seek(0, 0)
	reader = csv.NewReader(bufio.NewReaderSize(file, 64*1024))
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.Comma = c.detectDelimiter(inputPath)

	// Create PDF builder
	builder, err := pdf.NewBuilder(opts)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create PDF builder")
	}
	
	if c.onProgress != nil {
		builder.SetProgressCallback(c.onProgress)
	}

	builder.AddPage()

	// Read and write in chunks
	rowIndex := 0
	var headers []string
	style := pdf.DefaultStyle()
	headerStyle := pdf.HeaderStyle()
	rowHeight := style.FontSize + (style.Padding * 2) + 4

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if rowIndex == 0 && opts.HeaderRow {
			headers = record
			// Draw header row
			builder.SetFont(headerStyle.FontFamily, headerStyle.FontStyle, headerStyle.FontSize)
			builder.GetPdf().SetX(opts.Margin)
			for i, header := range headers {
				if i < len(colWidths) {
					builder.Cell(colWidths[i], rowHeight, header, headerStyle)
				}
			}
			builder.NewLine(rowHeight)
		} else {
			// Check for new page
			if builder.NeedsNewPage(rowHeight) {
				builder.AddPage()
				// Redraw headers on new page
				if opts.HeaderRow && len(headers) > 0 {
					builder.SetFont(headerStyle.FontFamily, headerStyle.FontStyle, headerStyle.FontSize)
					builder.GetPdf().SetX(opts.Margin)
					for i, header := range headers {
						if i < len(colWidths) {
							builder.Cell(colWidths[i], rowHeight, header, headerStyle)
						}
					}
					builder.NewLine(rowHeight)
				}
			}

			// Draw data row
			rowStyle := pdf.AlternatingRowStyle(rowIndex%2 == 0)
			builder.SetFont(style.FontFamily, style.FontStyle, style.FontSize)
			builder.GetPdf().SetX(opts.Margin)
			for i, cell := range record {
				if i < len(colWidths) {
					builder.Cell(colWidths[i], rowHeight, cell, rowStyle)
				}
			}
			builder.NewLine(rowHeight)
		}

		rowIndex++
	}

	return builder.Save(outputPath)
}
