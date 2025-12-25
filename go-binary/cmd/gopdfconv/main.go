package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nikunjkothiya/gopdfconv/internal/converter"
	"github.com/nikunjkothiya/gopdfconv/internal/pdf"
	"github.com/nikunjkothiya/gopdfconv/internal/worker"
	"github.com/nikunjkothiya/gopdfconv/pkg/errors"
)

// Version information
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
)

// Output format for Laravel parsing
type Output struct {
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	Error       *errors.ConversionError `json:"error,omitempty"`
	InputFile   string `json:"input_file,omitempty"`
	OutputFile  string `json:"output_file,omitempty"`
	Format      string `json:"format,omitempty"`
	ProcessTime int64  `json:"process_time_ms,omitempty"`
	FileSize    int64  `json:"file_size_bytes,omitempty"`
	PageCount   int    `json:"page_count,omitempty"`
}

func main() {
	// Define command-line flags
	inputFile := flag.String("input", "", "Input file path (CSV, XLSX, PPTX)")
	outputFile := flag.String("output", "", "Output PDF file path")
	formatFlag := flag.String("format", "auto", "Force input format (csv|xlsx|pptx|auto)")
	
	// Page options
	pageSize := flag.String("page-size", "A4", "Page size (A4|Letter|Legal|A3)")
	orientation := flag.String("orientation", "portrait", "Page orientation (portrait|landscape)")
	margin := flag.Float64("margin", 20, "Page margin in points")
	
	// Content options
	headerRow := flag.Bool("header", true, "Treat first row as header (CSV/Excel)")
	fontSize := flag.Float64("font-size", 10, "Base font size")
	headerText := flag.String("header-text", "", "Global header text (center)")
	footerText := flag.String("footer-text", "", "Global footer text (left)")

	// Advanced options
	customFont := flag.String("font", "", "Path to custom TTF font")
	watermarkText := flag.String("watermark-text", "", "Watermark text")
	watermarkImage := flag.String("watermark-image", "", "Path to watermark image")
	watermarkAlpha := flag.Float64("watermark-alpha", 0.2, "Watermark opacity (0.0-1.0)")

	// Smart Layout
	autoOrientation := flag.Bool("auto-orientation", true, "Automatically switch resolution if needed")
	
	// Styling options
	headerColor := flag.String("header-color", "", "Header background color (hex)")
	rowColor := flag.String("row-color", "", "Alternating row color (hex)")
	borderColor := flag.String("border-color", "", "Border color (hex)")
	gridLines := flag.Bool("grid-lines", true, "Show table grid lines")
	
	// Batch processing
	batchFiles := flag.String("batch", "", "Comma-separated list of input files")
	outputDir := flag.String("output-dir", "", "Output directory for batch processing")
	workers := flag.Int("workers", 0, "Number of parallel workers (0=auto)")
	
	// Other options
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	jsonOutput := flag.Bool("json", true, "Output results as JSON")
	version := flag.Bool("version", false, "Show version information")
	native := flag.Bool("native", false, "Force native Go conversion (skip LibreOffice)")
	libreOffice := flag.String("libreoffice", "", "Path to LibreOffice binary (for PPTX)")
	
	flag.Parse()
	
	// Handle version flag
	if *version {
		fmt.Printf("gopdfconv version %s (built %s)\n", Version, BuildTime)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
	
	// Build PDF options
	opts := pdf.DefaultOptions()
	opts.Margin = *margin
	opts.Margin = *margin
	opts.FontSize = *fontSize
	opts.HeaderRow = *headerRow
	// Advanced options
	opts.CustomFontPath = *customFont
	opts.WatermarkText = *watermarkText
	opts.WatermarkImage = *watermarkImage
	opts.WatermarkAlpha = *watermarkAlpha
	
	// Headers
	opts.HeaderText = *headerText
	opts.FooterText = *footerText
	opts.AutoOrientation = *autoOrientation
	
	// Styling options
	opts.HeaderColor = *headerColor
	opts.RowColor = *rowColor
	opts.BorderColor = *borderColor
	opts.ShowGridLines = *gridLines
	
	// Parse page size
	switch strings.ToLower(*pageSize) {
	case "a4":
		opts.PageSize = pdf.PageA4
	case "letter":
		opts.PageSize = pdf.PageLetter
	case "legal":
		opts.PageSize = pdf.PageLegal
	case "a3":
		opts.PageSize = pdf.PageA3
	case "tabloid":
		opts.PageSize = pdf.PageTabloid
	}
	
	// Parse orientation
	if strings.ToLower(*orientation) == "landscape" {
		opts.Orientation = pdf.Landscape
	} else {
		opts.Orientation = pdf.Portrait
	}
	
	// Handle batch processing
	if *batchFiles != "" {
		files := strings.Split(*batchFiles, ",")
		runBatchConversion(files, *outputDir, opts, *workers, *formatFlag, *libreOffice, *native, *jsonOutput, *verbose)
		return
	}
	
	// Validate single file arguments
	if *inputFile == "" {
		printError(errors.New(errors.ErrFileNotFound, "Input file is required"), *jsonOutput)
		flag.Usage()
		os.Exit(1)
	}
	
	if *outputFile == "" {
		// Auto-generate output filename
		base := strings.TrimSuffix(*inputFile, filepath.Ext(*inputFile))
		*outputFile = base + ".pdf"
	}
	
	// Run single conversion
	runSingleConversion(*inputFile, *outputFile, opts, *formatFlag, *libreOffice, *native, *jsonOutput, *verbose)
}

func runSingleConversion(inputPath, outputPath string, opts pdf.Options, formatFlag, libreOfficePath string, native, jsonOutput, verbose bool) {
	start := time.Now()
	
	// Progress callback
	progressCallback := func(percent int) {
		if jsonOutput {
			// Print progress to stderr to avoid polluting stdout JSON
			fmt.Fprintf(os.Stderr, "{\"progress\": %d}\n", percent)
		} else if verbose {
			fmt.Fprintf(os.Stderr, "\rProgress: %d%%", percent)
		}
	}
	
	// Detect format
	var format converter.FormatType
	if formatFlag == "auto" {
		format = converter.DetectFormat(inputPath)
	} else {
		format = converter.FormatType(formatFlag)
	}
	
	if verbose {
		fmt.Fprintf(os.Stderr, "Converting %s to %s (format: %s)\n", inputPath, outputPath, format)
	}
	
	var err error
	
	switch format {
	case converter.FormatCSV, converter.FormatTSV:
		csvConverter := converter.NewCSVConverter()
		csvConverter.SetProgressCallback(progressCallback)
		err = csvConverter.Convert(inputPath, outputPath, opts)
		
	case converter.FormatXLSX, converter.FormatXLSM, converter.FormatXLS:
		// For XLSX, try native first. For XLS, try LibreOffice first if available.
		if format == converter.FormatXLS {
			pptxConverter := converter.NewPPTXConverter()
			if libreOfficePath != "" {
				pptxConverter.SetLibreOfficePath(libreOfficePath)
			}
			
			// If we want native conversion for XLS, we must convert to XLSX first
			if native && pptxConverter.HasLibreOffice() {
				loConverter := converter.NewLibreOfficeConverter(pptxConverter.GetLibreOfficePath())
				tempXlsx := inputPath + ".xlsx"
				if err := loConverter.ConvertTo(inputPath, tempXlsx, "xlsx"); err == nil {
					defer os.Remove(tempXlsx)
					excelConverter := converter.NewExcelConverter()
			excelConverter.SetProgressCallback(progressCallback)
					err = excelConverter.Convert(tempXlsx, outputPath, opts)
				} else {
					// Fallback to direct LO conversion if temp conversion fails
					err = loConverter.Convert(inputPath, outputPath)
				}
			} else if pptxConverter.HasLibreOffice() && !native {
				loConverter := converter.NewLibreOfficeConverter(pptxConverter.GetLibreOfficePath())
				err = loConverter.Convert(inputPath, outputPath)
			} else {
				excelConverter := converter.NewExcelConverter()
			excelConverter.SetProgressCallback(progressCallback)
				err = excelConverter.Convert(inputPath, outputPath, opts)
			}
		} else {
			excelConverter := converter.NewExcelConverter()
			excelConverter.SetProgressCallback(progressCallback)
			err = excelConverter.Convert(inputPath, outputPath, opts)
		}
		
	case converter.FormatPPTX:
		pptxConverter := converter.NewPPTXConverter()
		if libreOfficePath != "" {
			pptxConverter.SetLibreOfficePath(libreOfficePath)
		}
		if native {
			pptxConverter.SetUseLibreOffice(false)
		}
		err = pptxConverter.Convert(inputPath, outputPath, opts)
		
	case converter.FormatPPT:
		// Check if LibreOffice is available for better fidelity
		pptxConverter := converter.NewPPTXConverter()
		if libreOfficePath != "" {
			pptxConverter.SetLibreOfficePath(libreOfficePath)
		}
		if pptxConverter.HasLibreOffice() && !native {
			// Use LibreOffice for best results
			loConverter := converter.NewLibreOfficeConverter(pptxConverter.GetLibreOfficePath())
			err = loConverter.Convert(inputPath, outputPath)
		} else {
			// Fall back to native PPT parser (text extraction only)
			pptConverter := converter.NewPPTConverter()
			err = pptConverter.Convert(inputPath, outputPath, opts)
		}
		
	default:
		err = errors.New(errors.ErrUnsupportedFormat, "Unsupported file format: "+string(format))
	}
	
	processTime := time.Since(start).Milliseconds()
	
	if err != nil {
		if convErr, ok := err.(*errors.ConversionError); ok {
			printError(convErr, jsonOutput)
		} else {
			printError(errors.Wrap(err, errors.ErrConversionFailed, "Conversion failed"), jsonOutput)
		}
		os.Exit(1)
	}
	
	// Get output file size
	var fileSize int64
	if info, statErr := os.Stat(outputPath); statErr == nil {
		fileSize = info.Size()
	}
	
	// Output success
	output := Output{
		Success:     true,
		Message:     "Conversion completed successfully",
		InputFile:   inputPath,
		OutputFile:  outputPath,
		Format:      string(format),
		ProcessTime: processTime,
		FileSize:    fileSize,
	}
	
	if jsonOutput {
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("✓ Converted %s to %s (%dms, %d bytes)\n", inputPath, outputPath, processTime, fileSize)
	}
}

func runBatchConversion(files []string, outputDir string, opts pdf.Options, numWorkers int, formatFlag, libreOfficePath string, native, jsonOutput, verbose bool) {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}
	
	// Create output directory if specified
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			printError(errors.Wrap(err, errors.ErrWriteFailed, "Failed to create output directory"), jsonOutput)
			os.Exit(1)
		}
	}
	
	// Build jobs
	var jobs []worker.Job
	for i, inputPath := range files {
		inputPath = strings.TrimSpace(inputPath)
		if inputPath == "" {
			continue
		}
		
		// Determine output path
		var outputPath string
		if outputDir != "" {
			base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
			outputPath = filepath.Join(outputDir, base+".pdf")
		} else {
			base := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
			outputPath = base + ".pdf"
		}
		
		// Detect format
		var format converter.FormatType
		if formatFlag == "auto" {
			format = converter.DetectFormat(inputPath)
		} else {
			format = converter.FormatType(formatFlag)
		}
		
		jobs = append(jobs, worker.Job{
			ID:         fmt.Sprintf("job-%d", i+1),
			InputPath:  inputPath,
			OutputPath: outputPath,
			Format:     format,
			Options:    opts,
		})
	}
	
	if len(jobs) == 0 {
		printError(errors.New(errors.ErrInvalidFormat, "No valid input files provided"), jsonOutput)
		os.Exit(1)
	}
	
	if verbose {
		fmt.Fprintf(os.Stderr, "Processing %d files with %d workers\n", len(jobs), numWorkers)
	}
	
	// Run batch conversion
	result := worker.RunBatch(jobs, numWorkers, libreOfficePath, native)
	
	if jsonOutput {
		fmt.Println(result.ToJSON())
	} else {
		fmt.Printf("Batch conversion complete:\n")
		fmt.Printf("  Total: %d files\n", result.TotalJobs)
		fmt.Printf("  Success: %d\n", result.Successful)
		fmt.Printf("  Failed: %d\n", result.Failed)
		fmt.Printf("  Time: %dms\n", result.TotalTime.Milliseconds())
	}
	
	if result.Failed > 0 {
		os.Exit(1)
	}
}

func printError(err *errors.ConversionError, jsonOutput bool) {
	if jsonOutput {
		output := Output{
			Success: false,
			Error:   err,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Message)
		if err.Details != "" {
			fmt.Fprintf(os.Stderr, "Details: %s\n", err.Details)
		}
	}
}
