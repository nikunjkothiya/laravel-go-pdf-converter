package converter

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nikunjkothiya/gopdfconv/pkg/errors"
)

// LibreOfficeConverter handles conversion using LibreOffice
type LibreOfficeConverter struct {
	libreOfficePath string
}

// NewLibreOfficeConverter creates a new LibreOffice converter
func NewLibreOfficeConverter(path string) *LibreOfficeConverter {
	return &LibreOfficeConverter{
		libreOfficePath: path,
	}
}

// Convert performs the conversion using LibreOffice
func (c *LibreOfficeConverter) Convert(inputPath, outputPath string) error {
	// Check if file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return errors.NewWithFile(errors.ErrFileNotFound, "File not found", inputPath)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return errors.Wrap(err, errors.ErrWriteFailed, "Failed to create output directory")
	}

	// Create a temp directory for LibreOffice output
	tempDir, err := os.MkdirTemp("", "gopdfconv-*")
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create temp directory")
	}
	defer os.RemoveAll(tempDir)

	// Run LibreOffice conversion with a temporary user profile and explicit Impress filter
	cmd := exec.Command(c.libreOfficePath,
		"-env:UserInstallation=file://"+tempDir+"/profile",
		"--headless",
		"--convert-to", "pdf:impress_pdf_Export",
		"--outdir", tempDir,
		inputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.NewWithDetails(errors.ErrConversionFailed, "LibreOffice conversion failed", inputPath, string(output))
	}

	// Find the generated PDF file in temp directory
	files, err := os.ReadDir(tempDir)
	if err != nil || len(files) == 0 {
		return errors.New(errors.ErrConversionFailed, "LibreOffice failed to generate PDF")
	}

	// Move the generated PDF to the final output path
	generatedPDF := filepath.Join(tempDir, files[0].Name())
	if err := os.Rename(generatedPDF, outputPath); err != nil {
		// If rename fails (e.g. across filesystems), try copy
		if err := copyFile(generatedPDF, outputPath); err != nil {
			return errors.Wrap(err, errors.ErrWriteFailed, "Failed to move generated PDF")
		}
	}

	return nil
}

// ConvertTo converts a file to a specific format using LibreOffice
func (c *LibreOfficeConverter) ConvertTo(inputPath, outputPath, format string) error {
	tempDir, err := os.MkdirTemp("", "gopdfconv-lo-*")
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to create temp directory")
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command(c.libreOfficePath,
		"-env:UserInstallation=file://"+tempDir+"/profile",
		"--headless",
		"--convert-to", format,
		"--outdir", tempDir,
		inputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.NewWithDetails(errors.ErrConversionFailed, "LibreOffice conversion failed", inputPath, string(output))
	}

	// Find the generated file in temp directory
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConversionFailed, "Failed to read temp directory")
	}

	var generatedFile string
	for _, f := range files {
		if !f.IsDir() && f.Name() != "profile" {
			generatedFile = filepath.Join(tempDir, f.Name())
			break
		}
	}

	if generatedFile == "" {
		return errors.New(errors.ErrConversionFailed, "LibreOffice failed to generate output file")
	}

	// Move to final destination
	if err := os.Rename(generatedFile, outputPath); err != nil {
		// If rename fails (e.g. cross-device), try copy
		input, err := os.ReadFile(generatedFile)
		if err != nil {
			return err
		}
		return os.WriteFile(outputPath, input, 0644)
	}

	return nil
}

// copyFile is a helper to copy a file if rename fails
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
