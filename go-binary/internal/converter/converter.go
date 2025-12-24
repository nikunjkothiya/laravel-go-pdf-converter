package converter

import (
	"github.com/nikunjkothiya/gopdfconv/internal/pdf"
)

// Converter is the interface for all file format converters
type Converter interface {
	// Convert performs the conversion from input to PDF output
	Convert(inputPath, outputPath string, opts pdf.Options) error

	// Validate checks if the input file is valid for this converter
	Validate(inputPath string) error

	// SupportedExtensions returns the file extensions this converter handles
	SupportedExtensions() []string
}

// Result represents the result of a conversion operation
type Result struct {
	Success     bool   `json:"success"`
	InputFile   string `json:"input_file"`
	OutputFile  string `json:"output_file"`
	Format      string `json:"format"`
	Pages       int    `json:"pages"`
	ProcessTime int64  `json:"process_time_ms"`
	FileSize    int64  `json:"file_size_bytes"`
	Error       string `json:"error,omitempty"`
}

// BatchResult represents the result of a batch conversion
type BatchResult struct {
	TotalFiles     int      `json:"total_files"`
	SuccessCount   int      `json:"success_count"`
	FailureCount   int      `json:"failure_count"`
	TotalTimeMs    int64    `json:"total_time_ms"`
	Results        []Result `json:"results"`
}

// FormatType represents the input file format
type FormatType string

const (
	FormatCSV   FormatType = "csv"
	FormatXLSX  FormatType = "xlsx"
	FormatXLS   FormatType = "xls"
	FormatPPTX  FormatType = "pptx"
	FormatPPT   FormatType = "ppt"
	FormatAuto  FormatType = "auto"
)

// DetectFormat determines the format from file extension
func DetectFormat(filename string) FormatType {
	ext := getExtension(filename)
	switch ext {
	case ".csv":
		return FormatCSV
	case ".xlsx":
		return FormatXLSX
	case ".xls":
		return FormatXLS
	case ".pptx":
		return FormatPPTX
	case ".ppt":
		return FormatPPT
	default:
		return FormatAuto
	}
}

// getExtension returns the lowercase file extension
func getExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext := filename[i:]
			// Convert to lowercase
			result := make([]byte, len(ext))
			for j, c := range ext {
				if c >= 'A' && c <= 'Z' {
					result[j] = byte(c + 32)
				} else {
					result[j] = byte(c)
				}
			}
			return string(result)
		}
	}
	return ""
}
