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
	maxRows int // Maximum rows to process (0 = unlimited)
	onProgress func(int)
}

// MaxRowsDefault is the default maximum number of rows to process
// This prevents memory issues and timeouts with very large files
const MaxRowsDefault = 10000

// NewExcelConverter creates a new Excel converter
func NewExcelConverter() *ExcelConverter {
	return &ExcelConverter{
		opts:    pdf.DefaultOptions(),
		maxRows: MaxRowsDefault,
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

// Convert performs the Excel to PDF conversion
func (c *ExcelConverter) Convert(inputPath, outputPath string, opts pdf.Options) error {
	// Validate input
	if err := c.Validate(inputPath); err != nil {
		return err
	}

	// Open Excel file
	f, err := excelize.OpenFile(inputPath)
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

	for sheetIndex, sheetName := range sheets {
		// Add new page for each sheet (except first)
		if sheetIndex > 0 {
			builder.AddPage()
		} else {
			builder.AddPage()
		}

		// Add sheet name as title
		titleStyle := pdf.HeaderStyle()
		titleStyle.FontSize = 14
		// Sheet title removed as per user request
		builder.NewLine(10)

		// Get all rows from the sheet
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue // Skip sheet on error
		}

		if len(rows) == 0 {
			continue // Skip empty sheets
		}

		// Apply row limit to prevent memory issues and timeouts
		truncated := false
		if c.maxRows > 0 && len(rows) > c.maxRows {
			rows = rows[:c.maxRows]
			truncated = true
		}

		// Calculate column widths (sample first 100 rows for performance)
		colWidths := c.calculateColumnWidths(rows, opts)

		// Prepare headers and data
		var headers []string
		var dataRows [][]string

		if opts.HeaderRow && len(rows) > 0 {
			headers = rows[0]
			// Center headers
			headerStyle := pdf.HeaderStyle()
			headerStyle.Alignment = pdf.AlignCenter
			
			if len(rows) > 1 {
				dataRows = rows[1:]
			}
		} else {
			dataRows = rows
		}

		// Add truncation notice if file was too large
		if truncated {
			dataRows = append(dataRows, []string{fmt.Sprintf("... (Showing first %d rows, file truncated for performance)", c.maxRows)})
		}

		// Draw the table
		if err := builder.DrawTable(headers, dataRows, colWidths); err != nil {
			return errors.Wrap(err, errors.ErrConversionFailed, "Failed to draw table")
		}
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

	// Open Excel file
	f, err := excelize.OpenFile(inputPath)
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

	for sheetIndex, sheetName := range sheetNames {
		// Verify sheet exists
		sheetIndex2, err := f.GetSheetIndex(sheetName)
		if err != nil || sheetIndex2 < 0 {
			continue // Skip non-existent sheets
		}

		// Add new page for each sheet (except first)
		if sheetIndex > 0 {
			builder.AddPage()
		} else {
			builder.AddPage()
		}

		// Add sheet name as title
		titleStyle := pdf.HeaderStyle()
		titleStyle.FontSize = 14
		// Sheet title removed as per user request
		builder.NewLine(10)

		// Get all rows from the sheet
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue
		}

		if len(rows) == 0 {
			builder.AddText("(Empty sheet)", pdf.DefaultStyle())
			continue
		}

		// Calculate column widths
		colWidths := c.calculateColumnWidths(rows, opts)

		// Prepare headers and data
		var headers []string
		var dataRows [][]string

		if opts.HeaderRow && len(rows) > 0 {
			headers = rows[0]
			if len(rows) > 1 {
				dataRows = rows[1:]
			}
		} else {
			dataRows = rows
		}

		// Draw the table
		if err := builder.DrawTable(headers, dataRows, colWidths); err != nil {
			return errors.Wrap(err, errors.ErrConversionFailed, "Failed to draw table")
		}
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

	// Apply min/max constraints - use reasonable minimums to prevent truncation
	const minColWidth = 40.0  // Minimum to show ~6 chars
	const maxColWidth = 180.0 // Maximum for any single column

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
