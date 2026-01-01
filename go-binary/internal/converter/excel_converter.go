package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nikunjkothiya/gopdfconv/internal/pdf"
	"github.com/nikunjkothiya/gopdfconv/pkg/errors"
	"github.com/xuri/excelize/v2"
)

// ExcelConverter handles Excel (XLSX/XLS) to PDF conversion
type ExcelConverter struct {
	opts    pdf.Options
	onProgress func(int)
}

// excelRowIterator adapts excelize.Rows to pdf.RowIterator interface
type excelRowIterator struct {
	rows *excelize.Rows
}

func (e *excelRowIterator) Next() bool {
	return e.rows.Next()
}

func (e *excelRowIterator) Columns() ([]string, error) {
	return e.rows.Columns()
}

// NewExcelConverter creates a new Excel converter
func NewExcelConverter() *ExcelConverter {
	return &ExcelConverter{
		opts:    pdf.DefaultOptions(),
	}
}

// SetProgressCallback sets the callback for progress reporting
func (c *ExcelConverter) SetProgressCallback(callback func(int)) {
	c.onProgress = callback
}

// SupportedExtensions returns extensions handled by this converter
func (c *ExcelConverter) SupportedExtensions() []string {
	return []string{".xlsx", ".xls", ".xlsm"}
}

// Validate checks if the input file is a valid Excel file
func (c *ExcelConverter) Validate(inputPath string) error {
	// Check file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return errors.NewWithFile(errors.ErrFileNotFound, "File not found", inputPath)
	}

	// Try to open as Excel file
	f, err := excelize.OpenFile(inputPath)
	if err != nil {
		return errors.NewWithDetails(errors.ErrInvalidFormat, "Invalid Excel format", inputPath, err.Error())
	}
	defer f.Close()

	// Check for at least one sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return errors.NewWithFile(errors.ErrInvalidFormat, "Excel file has no sheets", inputPath)
	}

	return nil
}

// Convert performs the Excel to PDF conversion using streaming for large files
func (c *ExcelConverter) Convert(inputPath, outputPath string, opts pdf.Options) error {
	// Validate input
	if err := c.Validate(inputPath); err != nil {
		return err
	}

	// Open Excel file with memory optimization options
	f, err := excelize.OpenFile(inputPath, excelize.Options{
		UnzipSizeLimit: 100 << 20, // 100MB limit
		UnzipXMLSizeLimit: 50 << 20, // 50MB XML limit
	})
	if err != nil {
		return errors.NewWithDetails(errors.ErrConversionFailed, "Failed to open Excel file", inputPath, err.Error())
	}
	defer f.Close()

	// Create PDF builder
	builder, err := pdf.NewBuilder(opts)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create PDF builder")
	}
	
	if c.onProgress != nil {
		builder.SetProgressCallback(c.onProgress)
	}

	// Get all sheets
	sheets := f.GetSheetList()

	for _, sheetName := range sheets {
		// Add new page for each sheet
		builder.AddPage()

		// Add sheet name as title
		builder.NewLine(10)

		// Use streaming reader for large files to avoid memory issues
		streamRows, err := f.Rows(sheetName)
		if err != nil {
			continue // Skip sheet on error
		}
		
		// First pass: sample rows for column width calculation (memory efficient)
		var sampleRows [][]string
		rowCount := 0
		for streamRows.Next() && rowCount < 100 {
			row, err := streamRows.Columns()
			if err != nil {
				continue
			}
			sampleRows = append(sampleRows, row)
			rowCount++
		}
		streamRows.Close()

		if len(sampleRows) == 0 {
			continue // Skip empty sheets
		}

		// Calculate column widths from sample
		colWidths := c.calculateColumnWidths(sampleRows, opts)

		// Prepare headers
		var headers []string
		if opts.HeaderRow && len(sampleRows) > 0 {
			headers = sampleRows[0]
		}

		// Second pass: stream rows directly to PDF (memory efficient)
		streamRows, err = f.Rows(sheetName)
		if err != nil {
			continue
		}

		// Draw table with streaming using adapter
		rowIterator := &excelRowIterator{rows: streamRows}
		if err := builder.DrawTableStreaming(headers, rowIterator, colWidths, opts.HeaderRow); err != nil {
			streamRows.Close()
			return errors.Wrap(err, errors.ErrConversionFailed, "Failed to draw table")
		}
		streamRows.Close()
	}

	// Save the PDF
	if err := builder.Save(outputPath); err != nil {
		return errors.Wrap(err, errors.ErrWriteFailed, "Failed to save PDF")
	}

	return nil
}

// ConvertWithOptions allows specific sheet selection and other options
func (c *ExcelConverter) ConvertWithOptions(inputPath, outputPath string, opts pdf.Options, sheetNames []string) error {
	// Validate input
	if err := c.Validate(inputPath); err != nil {
		return err
	}

	// Open Excel file with memory optimization
	f, err := excelize.OpenFile(inputPath, excelize.Options{
		UnzipSizeLimit: 100 << 20,
		UnzipXMLSizeLimit: 50 << 20,
	})
	if err != nil {
		return errors.NewWithDetails(errors.ErrConversionFailed, "Failed to open Excel file", inputPath, err.Error())
	}
	defer f.Close()

	// Create PDF builder
	builder, err := pdf.NewBuilder(opts)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create PDF builder")
	}
	
	if c.onProgress != nil {
		builder.SetProgressCallback(c.onProgress)
	}

	// If no sheets specified, use all sheets
	if len(sheetNames) == 0 {
		sheetNames = f.GetSheetList()
	}

	for sheetIdx, sheetName := range sheetNames {
		// Verify sheet exists
		sheetIndex, err := f.GetSheetIndex(sheetName)
		if err != nil || sheetIndex < 0 {
			continue // Skip non-existent sheets
		}

		// Add new page for each sheet (except first)
		if sheetIdx > 0 {
			builder.AddPage()
		} else {
			builder.AddPage()
		}

		builder.NewLine(10)

		// Use streaming reader - sample first for column widths
		streamRows, err := f.Rows(sheetName)
		if err != nil {
			continue
		}
		
		var sampleRows [][]string
		rowCount := 0
		for streamRows.Next() && rowCount < 100 {
			row, err := streamRows.Columns()
			if err != nil {
				continue
			}
			sampleRows = append(sampleRows, row)
			rowCount++
		}
		streamRows.Close()

		if len(sampleRows) == 0 {
			builder.AddText("(Empty sheet)", pdf.DefaultStyle())
			continue
		}

		// Calculate column widths
		colWidths := c.calculateColumnWidths(sampleRows, opts)

		// Prepare headers
		var headers []string
		if opts.HeaderRow && len(sampleRows) > 0 {
			headers = sampleRows[0]
		}

		// Second pass: stream to PDF
		streamRows, err = f.Rows(sheetName)
		if err != nil {
			continue
		}

		// Use adapter for streaming
		rowIterator := &excelRowIterator{rows: streamRows}
		if err := builder.DrawTableStreaming(headers, rowIterator, colWidths, opts.HeaderRow); err != nil {
			streamRows.Close()
			return errors.Wrap(err, errors.ErrConversionFailed, "Failed to draw table")
		}
		streamRows.Close()
	}

	// Save the PDF
	if err := builder.Save(outputPath); err != nil {
		return errors.Wrap(err, errors.ErrWriteFailed, "Failed to save PDF")
	}

	return nil
}

// ConvertSheetToCSV exports a specific sheet to CSV, then converts to PDF
func (c *ExcelConverter) ConvertSheetToCSV(inputPath, outputPath, sheetName string) (string, error) {
	f, err := excelize.OpenFile(inputPath)
	if err != nil {
		return "", errors.NewWithDetails(errors.ErrConversionFailed, "Failed to open Excel file", inputPath, err.Error())
	}
	defer f.Close()

	// Get rows from specified sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return "", errors.NewWithDetails(errors.ErrConversionFailed, "Failed to read sheet", sheetName, err.Error())
	}

	// Create temp CSV file
	csvPath := filepath.Join(filepath.Dir(outputPath), fmt.Sprintf("%s_%s.csv", 
		strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)), 
		sheetName))

	csvFile, err := os.Create(csvPath)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrWriteFailed, "Failed to create temp CSV")
	}
	defer csvFile.Close()

	// Write rows to CSV
	for _, row := range rows {
		line := strings.Join(row, ",")
		csvFile.WriteString(line + "\n")
	}

	return csvPath, nil
}

// calculateColumnWidths calculates optimal column widths based on content
func (c *ExcelConverter) calculateColumnWidths(rows [][]string, opts pdf.Options) []float64 {
	if len(rows) == 0 {
		return nil
	}

	// Find the maximum number of columns
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	if maxCols == 0 {
		return nil
	}

	// Calculate max width for each column
	colMaxWidths := make([]float64, maxCols)

	// Sample first 100 rows for width calculation
	sampleSize := 100
	if len(rows) < sampleSize {
		sampleSize = len(rows)
	}

	for i := 0; i < sampleSize; i++ {
		row := rows[i]
		for j, cell := range row {
			if j >= maxCols {
				continue
			}
			// Estimate width: ~6 points per character + padding
			width := float64(len(cell))*6 + 8
			if width > colMaxWidths[j] {
				colMaxWidths[j] = width
			}
		}
	}

	// Use custom min/max from options, or defaults
	minColWidth := opts.MinColumnWidth
	maxColWidth := opts.MaxColumnWidth
	if minColWidth <= 0 {
		minColWidth = 40.0
	}
	if maxColWidth <= 0 {
		maxColWidth = 180.0
	}

	for i := range colMaxWidths {
		if colMaxWidths[i] < minColWidth {
			colMaxWidths[i] = minColWidth
		}
		if colMaxWidths[i] > maxColWidth {
			colMaxWidths[i] = maxColWidth
		}
	}

	// Scale to fit page width
	totalWidth := 0.0
	for _, w := range colMaxWidths {
		totalWidth += w
	}

	contentWidth := opts.ContentWidth()
	if totalWidth > contentWidth {
		scale := contentWidth / totalWidth
		for i := range colMaxWidths {
			colMaxWidths[i] *= scale
			// Ensure minimum readable width (at least 35 points = ~5-6 chars)
			if colMaxWidths[i] < 35 {
				colMaxWidths[i] = 35
			}
		}
	}

	return colMaxWidths
}

// GetSheetList returns all sheet names in an Excel file
func GetSheetList(inputPath string) ([]string, error) {
	f, err := excelize.OpenFile(inputPath)
	if err != nil {
		return nil, errors.NewWithDetails(errors.ErrConversionFailed, "Failed to open Excel file", inputPath, err.Error())
	}
	defer f.Close()

	return f.GetSheetList(), nil
}
