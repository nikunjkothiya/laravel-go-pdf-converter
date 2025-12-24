# Laravel Go PDF Converter

**High-performance CSV, Excel, and PowerPoint to PDF converter for Laravel, powered by Go for maximum speed and minimal resource usage.**

- 🚀 **50-70% faster** than pure PHP solutions  
- 💾 **Low memory footprint** – handles large files without exhausting memory  
- 🔧 **Zero external dependencies** – no LibreOffice required for most conversions  
- ⚡ **Queue-ready** – dispatch conversions to background jobs  
- 🎨 **Professional Table Rendering** – grid lines, smart alignment, and centered layouts for Excel/CSV  
- 🌈 **Smart Color Fallback** – automatically fixes white-on-white text in PowerPoint conversions  
- 📝 **Global Headers & Footers** – add custom text and mandatory "Page X of Y" numbering

---

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [Facade API](#facade-api)
  - [CSV Conversion](#csv-conversion)
  - [Excel Conversion](#excel-conversion)
  - [PowerPoint Conversion](#powerpoint-conversion)
  - [Batch Processing](#batch-processing)
  - [Queue Jobs](#queue-jobs)
  - [Artisan Commands](#artisan-commands)
- [Configuration](#configuration)
- [Performance](#performance)
- [Supported Formats](#supported-formats)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

---

## Features

| Format | Input Extensions | Features |
|--------|-----------------|----------|
| **CSV** | `.csv`, `.tsv` | Auto-detect delimiter, streaming for large files, header styling |
| **Excel** | `.xlsx`, `.xls`, `.xlsm` | Professional grid lines, smart alignment (numbers/text), centered tables |
| **PowerPoint** | `.pptx`, `.ppt` | Slide extraction, Smart Color fallback for text visibility, optional LibreOffice |

### Why Go?

- **Memory Efficient**: Processes files in streams, doesn't load entire file into memory
- **Concurrent**: Batch processing uses goroutines for parallel conversion
- **Fast**: Native binary execution, no interpreter overhead
- **Cross-platform**: Pre-built binaries for Linux, macOS, and Windows

---

## Requirements

- PHP 8.1+
- Laravel 10.x or 11.x
- Go 1.21+ (only for building from source)

---

## Installation

### Step 1: Install the Package

```bash
composer require nikunjkothiya/laravel-go-pdf-converter
```

### Step 2: Publish Configuration (Optional)

```bash
php artisan vendor:publish --tag=gopdf-config
```

### Step 3: Install the Go Binary

```bash
php artisan gopdf:install
```

This will automatically:
1. ✅ Download the correct binary for your platform (Linux/macOS/Windows)
2. ✅ Check for LibreOffice installation
3. ✅ Offer to install LibreOffice for full PPT/PPTX support

#### Optional: Install with LibreOffice

```bash
# Interactive installation with LibreOffice prompt
php artisan gopdf:install

# Auto-install LibreOffice (requires sudo on Linux)
php artisan gopdf:install --with-libreoffice

# Skip LibreOffice prompt
php artisan gopdf:install --skip-libreoffice
```

---

## Quick Start

```php
use NikunjKothiya\GoPdfConverter\Facades\PdfConverter;

// Convert CSV to PDF
PdfConverter::csv('data.csv')->toPdf('output.pdf')->convert();

// Convert Excel to PDF
PdfConverter::excel('report.xlsx')->toPdf('report.pdf')->convert();

// Convert PowerPoint to PDF
PdfConverter::pptx('presentation.pptx')->toPdf('slides.pdf')->convert();
```

---

## Usage

### Facade API

The `PdfConverter` facade provides a fluent, chainable API:

```php
use NikunjKothiya\GoPdfConverter\Facades\PdfConverter;

// Basic conversion
$result = PdfConverter::from('input.csv')
    ->toPdf('output.pdf')
    ->convert();

// Returns array with conversion details
// [
//     'success' => true,
//     'input_file' => '/path/to/input.csv',
//     'output_file' => '/path/to/output.pdf',
//     'format' => 'csv',
//     'process_time_ms' => 234,
//     'file_size_bytes' => 45678,
// ]

// Global Headers & Footers
PdfConverter::from('input.xlsx')
    ->headerText('Confidential Report')
    ->footerText('© 2025 Nikunj Kothiya')
    ->convert();
// Note: "Page X of Y" is automatically added to the footer.
```

### CSV Conversion

```php
// Simple conversion
PdfConverter::csv('data.csv')->toPdf('output.pdf')->convert();

// Handling Wide CSVs (many columns)
// This uses A3 landscape, smaller font, and tight margins for best fit
PdfConverter::csv('wide_data.csv')
    ->wideFormat()
    ->convert();

// Custom Page Sizes
PdfConverter::csv('data.csv')
    ->tabloid()               // 11x17 inches (Great for huge tables)
    ->a3()                    // 297x420 mm
    ->landscape()
    ->convert();

// With options
PdfConverter::csv('data.csv')
    ->toPdf('output.pdf')
    ->pageSize('Letter')      // A4, Letter, Legal, A3, Tabloid
    ->landscape()             // Or: ->portrait()
    ->withHeaders()           // Style first row as header
    ->fontSize(11)            // Base font size
    ->margin(25)              // Page margin in points
    ->convert();
```

### Path & Storage Handling

The package integrates seamlessly with Laravel's path and storage systems:

```php
// 1. Relative to Project Root (Default)
PdfConverter::csv('storage/app/data.csv')->convert();

// 2. Using Laravel Storage Disks
PdfConverter::csv('reports/data.csv')
    ->disk('s3')              // Uses S3 disk for both input and output
    ->toPdf('exports/data.pdf')
    ->convert();

// 3. Absolute Paths
PdfConverter::csv('/var/www/html/storage/app/data.csv')->convert();
```

> [!TIP]
> When using `->disk('name')`, the package will automatically resolve the full local path using `Storage::disk('name')->path($path)`. This is perfect for local, public, or mounted disks.

### Excel Conversion

```php
// Convert all sheets
PdfConverter::excel('workbook.xlsx')
    ->toPdf('output.pdf')
    ->convert();

// With professional styling
PdfConverter::xlsx('report.xlsx')
    ->toPdf('report.pdf')
    ->a4()
    ->portrait()
    ->withHeaders()
    ->convert();

// The native Excel renderer provides:
// - Visible grid lines (borders)
// - Smart alignment (numbers right-aligned, text left-aligned)
// - Automatic table centering for balanced margins
// - Automatic bridge for legacy .xls files (via LibreOffice)
```

### PowerPoint Conversion

```php
// Basic conversion
PdfConverter::pptx('presentation.pptx')
    ->toPdf('slides.pdf')
    ->convert();

// The converter will:
// 1. Try native Go extraction with Smart Color (fixes invisible text)
// 2. Fall back to LibreOffice if installed (for full fidelity)

// Force native conversion (useful if backgrounds don't render correctly)
PdfConverter::pptx('presentation.pptx')
    ->native()
    ->convert();
```

### Batch Processing

Convert multiple files at once with parallel processing:

```php
// Batch convert files
$result = PdfConverter::batch([
    'file1.csv',
    'file2.xlsx',
    'file3.pptx',
])
->outputDir('/path/to/output')
->workers(4)              // Parallel workers (default: CPU cores)
->landscape()             // Apply to all files
->convert();

// Returns summary
// [
//     'total_jobs' => 3,
//     'successful' => 3,
//     'failed' => 0,
//     'total_time_ns' => 1234567890,
//     'results' => [...],
// ]
```

### Queue Jobs

Dispatch conversions to Laravel queues for background processing:

```php
// Dispatch to queue
PdfConverter::csv('data.csv')
    ->toPdf('output.pdf')
    ->queue();

// Specify connection and queue name
PdfConverter::csv('data.csv')
    ->toPdf('output.pdf')
    ->queue('redis', 'pdf-conversions');

// Batch queue
PdfConverter::batch($files)
    ->outputDir($outputDir)
    ->queue();
```

The jobs include:
- Automatic retry (3 attempts by default)
- Exponential backoff
- Horizon tags for monitoring
- Proper logging

### Artisan Commands

#### Single File Conversion

```bash
# Basic conversion
php artisan pdf:convert input.csv output.pdf

# With options
php artisan pdf:convert input.xlsx output.pdf \
    --page-size=Letter \
    --landscape \
    --margin=25 \
    --font-size=11 \
    --native \
    --header-text="My Report" \
    --footer-text="Copyright 2025"
```

# Async via queue
php artisan pdf:convert input.csv output.pdf --queue

#### Install/Update Binary

```bash
# Install binary for current platform
php artisan gopdf:install

# Force reinstall
php artisan gopdf:install --force

# Custom install path
php artisan gopdf:install --path=/usr/local/bin
```

---

## Configuration

After publishing the config file (`config/gopdf.php`):

```php
return [
    // Custom binary path (default: auto-detect)
    'binary_path' => env('GOPDF_BINARY_PATH', null),

    // LibreOffice path for PPTX (default: auto-detect)
    'libreoffice_path' => env('GOPDF_LIBREOFFICE_PATH', null),

    // Default page settings
    'defaults' => [
        'page_size' => 'A4',
        'orientation' => 'portrait',
        'margin' => 20,
        'font_size' => 10,
        'header_row' => true,
    ],

    // Timeout settings (seconds)
    'timeout' => [
        'single' => 120,
        'batch' => 600,
    ],

    // Queue settings
    'queue' => [
        'connection' => null,  // Use default
        'queue' => 'pdf-conversions',
        'tries' => 3,
        'backoff' => 30,
    ],
];
```

### Environment Variables

```env
# Custom binary path
GOPDF_BINARY_PATH=/path/to/gopdfconv

# LibreOffice for PPTX
GOPDF_LIBREOFFICE_PATH=/usr/bin/libreoffice

# Queue configuration
GOPDF_QUEUE_CONNECTION=redis
GOPDF_QUEUE_NAME=pdf-conversions

# Batch processing
GOPDF_BATCH_WORKERS=4

# Logging
GOPDF_LOGGING=true
GOPDF_LOG_CHANNEL=stack
```

---

## Performance

### Benchmarks

Tested on standard hardware (2-core CPU, 8GB RAM) using samples files:

| File Type | File Size | This Package | Notes |
|-----------|-----------|--------------|-------|
| **CSV** | 14 MB | ~21s | Large data set conversion |
| **PPTX** | 15 MB | ~5s | Native Go conversion |
| **PPT** | 2 MB | ~1.6s | Legacy PowerPoint |
| **Excel** | 0.5 MB | ~1.2s | Native XLSX conversion |
| **CSV** | 1 KB | ~10ms | Small file overhead |

### Memory Usage

| Operation | This Package | Pure PHP |
|-----------|--------------|----------|
| 10 MB CSV | ~50 MB | ~500 MB |
| 100 MB CSV | ~80 MB | Memory exhausted |

---

## Supported Formats

| Format | Read | Notes |
|--------|------|-------|
| CSV | ✅ | Auto-detect delimiter (comma, tab, semicolon, pipe) |
| TSV | ✅ | Tab-separated values |
| XLSX | ✅ | Excel 2007+ format |
| XLS | ✅ | Legacy Excel format |
| XLSM | ✅ | Excel with macros |
| PPTX | ✅ | PowerPoint 2007+ format (native Go, or LibreOffice for full fidelity) |
| PPT | ✅ | Legacy PowerPoint (text extraction native, full fidelity with LibreOffice) |
| ODP | ⚠️ | Requires LibreOffice |

> **Note**: For PPT/PPTX files, the package will:
> 1. **With LibreOffice**: Full visual fidelity conversion (layouts, backgrounds, images)
> 2. **Without LibreOffice (Native)**: Text extraction with **Smart Color** (automatically converts white text to black for visibility). Use the `--native` flag to force this mode.

---

## Error Handling

```php
use NikunjKothiya\GoPdfConverter\Facades\PdfConverter;
use NikunjKothiya\GoPdfConverter\Exceptions\PdfConversionException;
use NikunjKothiya\GoPdfConverter\Exceptions\FileNotFoundException;
use NikunjKothiya\GoPdfConverter\Exceptions\UnsupportedFormatException;

try {
    PdfConverter::csv('data.csv')->toPdf('output.pdf')->convert();
} catch (FileNotFoundException $e) {
    // Input file doesn't exist
    echo "File not found: " . $e->getInputFile();
} catch (UnsupportedFormatException $e) {
    // Unsupported file type
    echo "Format not supported: " . $e->getMessage();
} catch (PdfConversionException $e) {
    // General conversion error
    echo "Error: " . $e->getMessage();
    echo "Code: " . $e->getErrorCode();
    echo "Details: " . $e->getDetails();
}
```

---

## Troubleshooting

### Binary Not Found

```bash
# Check if binary is installed
php artisan gopdf:install

# Or specify path in .env
GOPDF_BINARY_PATH=/path/to/gopdfconv
```

### Permission Denied

```bash
# Make binary executable
chmod +x vendor/nikunjkothiya/laravel-go-pdf-converter/bin/gopdfconv-*
```

### PPTX Low Quality

For best PowerPoint fidelity, install LibreOffice:

```bash
# Ubuntu/Debian
sudo apt-get install libreoffice

# macOS
brew install libreoffice

# Then set in .env
GOPDF_LIBREOFFICE_PATH=/usr/bin/libreoffice
```

### Large File Timeout

```php
// Increase timeout for large files
PdfConverter::csv('huge.csv')
    ->timeout(600)  // 10 minutes
    ->convert();
```

---

## Building from Source

If pre-built binaries don't work for your platform:

```bash
# Navigate to go-binary directory
cd vendor/nikunjkothiya/laravel-go-pdf-converter/go-binary

# Install dependencies
go mod download

# Build for current platform
make build

# Or build for all platforms
make build-all
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

### Third-Party Dependencies

This package uses the following open-source libraries, all with permissive licenses:

| Library | License | Purpose |
|---------|---------|---------|
| [signintech/gopdf](https://github.com/signintech/gopdf) | MIT | PDF generation |
| [xuri/excelize](https://github.com/xuri/excelize) | BSD 3-Clause | Excel parsing |
| [richardlehane/mscfb](https://github.com/richardlehane/mscfb) | Apache 2.0 | Legacy PPT parsing |
| [phpdave11/gofpdi](https://github.com/phpdave11/gofpdi) | MIT | PDF import |
| golang.org/x/* | BSD 3-Clause | Go standard extensions |

**All licenses allow**: ✅ Commercial use, ✅ Modification, ✅ Distribution

---

## Support

- ⭐ Star the repository on GitHub
- 🐛 Report bugs via GitHub Issues
- 💡 Suggest features via GitHub Discussions

## 💖 Support the Project

If you find this project helpful and want to support its development, you can buy me a coffee!

[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-Support%20Me-orange.svg?style=flat-square&logo=buy-me-a-coffee)](https://buymeacoffee.com/nikunjkothiya)

<img src="https://raw.githubusercontent.com/nikunjkothiya/assets/main/qr-code.png"
     alt="Buy Me A Coffee QR Code"
     width="180" />
     
---

## 📄 License

This package is open-source software licensed under the **MIT license**.

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/nikunjkothiya">Nikunj Kothiya</a>
</p>
