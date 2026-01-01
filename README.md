# Laravel Go PDF Converter

**High-performance CSV, Excel, and PowerPoint to PDF converter for Laravel, powered by Go for maximum speed and minimal resource usage.**

- 🚀 **Fast conversion** using native Go binary
- 💾 **Low memory footprint** – handles large files efficiently
- ⚡ **Queue-ready** – dispatch conversions to background jobs
- 🎨 **Professional Table Rendering** – grid lines, smart alignment, and centered layouts
- 📝 **Text Wrapping** – long content automatically wraps to multiple lines
- 📄 **Headers & Footers** – custom text with automatic page numbering

---

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [Table Styling & Customization](#table-styling--customization-excelcsv)
- [Supported Formats](#supported-formats)
- [Configuration](#configuration)
- [Error Handling](#error-handling)
- [Troubleshooting](#troubleshooting)
- [License](#license)

---

## Features

| Format           | Input Extensions | Conversion Method              |
| ---------------- | ---------------- | ------------------------------ |
| **CSV**          | `.csv`, `.tsv`   | Native Go (no dependencies)    |
| **Excel**        | `.xlsx`, `.xlsm` | Native Go (no dependencies)    |
| **Excel Legacy** | `.xls`           | LibreOffice → XLSX → Native Go |
| **PowerPoint**   | `.pptx`, `.ppt`  | LibreOffice (full fidelity)    |

### Key Features

- **Native Go Binary**: Pre-compiled binaries for Linux, macOS, and Windows included
- **No Go Installation Required**: Just install the PHP package and it works
- **LibreOffice Integration**: For PowerPoint and legacy Excel files with full visual fidelity
- **Queue Support**: Built-in Laravel queue jobs with retry logic
- **Batch Processing**: Convert multiple files in parallel
- **🎨 Full Table Styling**: Customize colors, row heights, column widths, fonts, and more
- **📊 Professional Output**: Grid lines, smart alignment, text wrapping, and centered layouts

---

## Requirements

- PHP 8.1+
- Laravel 10.x, 11.x, or 12.x
- **LibreOffice** (required for PPT, PPTX, and XLS files)

### LibreOffice Installation

LibreOffice is **required** for PowerPoint and legacy Excel conversions:

```bash
# Ubuntu/Debian
sudo apt-get install libreoffice

# macOS
brew install --cask libreoffice

# Windows
# Download from https://www.libreoffice.org/download/download/
```

---

## Installation

```bash
composer require nikunjkothiya/laravel-go-pdf-converter
```

### Publish Configuration (Optional)

```bash
php artisan vendor:publish --tag=gopdf-config
```

---

## Quick Start

```php
use NikunjKothiya\GoPdfConverter\Facades\PdfConverter;

// Convert CSV to PDF
PdfConverter::csv('data.csv')->toPdf('output.pdf')->convert();

// Convert Excel to PDF
PdfConverter::excel('report.xlsx')->toPdf('report.pdf')->convert();

// Convert PowerPoint to PDF (requires LibreOffice)
PdfConverter::pptx('presentation.pptx')->toPdf('slides.pdf')->convert();
```

---

## Usage

### Facade API

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
```

### CSV Conversion

```php
// Simple conversion
PdfConverter::csv('data.csv')->toPdf('output.pdf')->convert();

// With options
PdfConverter::csv('data.csv')
    ->toPdf('output.pdf')
    ->pageSize('A4')          // A4, Letter, Legal, A3, Tabloid
    ->landscape()             // Or: ->portrait()
    ->withHeaders()           // Style first row as header
    ->fontSize(11)            // Base font size
    ->margin(25)              // Page margin in points
    ->convert();

// Wide format for many columns
PdfConverter::csv('wide_data.csv')
    ->wideFormat()            // A3 landscape with smaller font
    ->convert();
```

### Excel Conversion

```php
// XLSX files - native conversion
PdfConverter::excel('workbook.xlsx')
    ->toPdf('output.pdf')
    ->convert();

// XLS files - converts to XLSX first via LibreOffice
PdfConverter::excel('legacy.xls')
    ->toPdf('output.pdf')
    ->convert();

// With styling options
PdfConverter::xlsx('report.xlsx')
    ->toPdf('report.pdf')
    ->landscape()
    ->withHeaders()
    ->headerText('Confidential Report')
    ->footerText('© 2025 Company')
    ->convert();
```

### PowerPoint Conversion

PowerPoint files require LibreOffice for full visual fidelity (backgrounds, images, layouts).

> **Note:** Table styling options (colors, row heights, column widths, etc.) do **not** apply to PowerPoint conversions. PowerPoint files use slide-based rendering. Only general options are supported: page size, orientation, margins, watermark, header/footer text, and custom fonts.

```php
// PPTX conversion (requires LibreOffice)
PdfConverter::pptx('presentation.pptx')
    ->toPdf('slides.pdf')
    ->convert();

// PPT legacy format (requires LibreOffice)
PdfConverter::pptx('old_presentation.ppt')
    ->toPdf('slides.pdf')
    ->convert();

// Force native mode (text extraction only, no backgrounds/images)
PdfConverter::pptx('presentation.pptx')
    ->native()
    ->convert();

// Supported options for PowerPoint (general options only)
PdfConverter::pptx('presentation.pptx')
    ->toPdf('slides.pdf')
    ->pageSize('A4')              // Page size
    ->landscape()                 // Orientation
    ->margin(25)                  // Page margins
    ->watermarkText('DRAFT')      // Watermark
    ->headerText('Company Name')  // Page header
    ->footerText('Confidential')  // Page footer
    ->convert();
```

### Headers & Footers

Add custom headers and footers to your PDF documents:

```php
PdfConverter::csv('data.csv')
    ->headerText('Quarterly Report')    // Custom text centered at top of each page
    ->footerText('Generated by System') // Custom text at bottom left of each page
    ->convert();
```

**Header Behavior:**
- Header text is displayed centered at the top of each page
- Only shown when `headerText()` is provided (no default header)
- Uses black text with 10pt font size for clear visibility

**Footer Behavior:**
- Custom footer text appears on the left side
- Page numbers ("Page X of Y") are automatically added on the right side
- Footer is always displayed with page numbering

### Table Styling & Customization (Excel/CSV Only)

Customize the appearance of tables when converting Excel or CSV files.

> **Note:** Table styling options (colors, row heights, column widths, cell padding, font styling, grid lines) **only apply to spreadsheet formats** (CSV, XLS, XLSX). PowerPoint files (PPT, PPTX) use their own slide-based rendering and ignore these table-specific options. For PowerPoint, only general options apply: page size, orientation, margins, watermark, header/footer text, and custom fonts.

#### Row & Cell Sizing

```php
PdfConverter::excel('data.xlsx')
    ->rowHeight(25)           // Custom row height in points (0 = auto)
    ->headerHeight(30)        // Custom header row height (0 = auto)
    ->cellPadding(6)          // Cell padding in points (default: 4)
    ->minColumnWidth(50)      // Minimum column width (default: 40)
    ->maxColumnWidth(200)     // Maximum column width (default: 180)
    ->convert();
```

#### Color Customization

```php
PdfConverter::excel('report.xlsx')
    ->headerColor('4A90D9')       // Header background color (hex without #)
    ->headerTextColor('FFFFFF')   // Header text color
    ->rowColor('F5F5F5')          // Alternating row background color
    ->rowTextColor('333333')      // Row text color
    ->borderColor('CCCCCC')       // Table border color
    ->showGridLines(true)         // Show/hide grid lines (default: true)
    ->convert();
```

#### Font Styling

```php
PdfConverter::csv('data.csv')
    ->headerFontSize(12)      // Header font size (0 = auto, uses FontSize + 1)
    ->headerBold(true)        // Make header text bold (default: true)
    ->fontSize(10)            // Base font size for data rows
    ->convert();
```

#### Complete Styling Example

```php
PdfConverter::excel('sales_report.xlsx')
    ->toPdf('styled_report.pdf')
    ->landscape()
    ->a4()
    // Row sizing
    ->rowHeight(22)
    ->headerHeight(28)
    ->cellPadding(5)
    // Colors
    ->headerColor('2C3E50')       // Dark blue table header background
    ->headerTextColor('FFFFFF')   // White table header text
    ->rowColor('ECF0F1')          // Light gray alternating rows
    ->rowTextColor('2C3E50')      // Dark text
    ->borderColor('BDC3C7')       // Gray borders
    // Font
    ->headerFontSize(11)
    ->headerBold(true)
    ->fontSize(9)
    // Grid
    ->showGridLines(true)
    // Page Headers/Footers (optional - only shown if provided)
    ->headerText('Sales Report - Q4 2025')  // Centered at top of each page
    ->footerText('Confidential')            // Left side of footer
    ->convert();
```

**Note:** The `headerText()` method sets the page header (document title at top), while `headerColor()`, `headerTextColor()`, etc. style the table's first row header.

#### CLI Options for Styling

All styling options are also available via command line:

```bash
php artisan pdf:convert data.xlsx report.pdf \
    --row-height=25 \
    --header-height=30 \
    --cell-padding=6 \
    --min-col-width=50 \
    --max-col-width=200 \
    --header-color=4A90D9 \
    --header-text-color=FFFFFF \
    --row-color=F5F5F5 \
    --row-text-color=333333 \
    --border-color=CCCCCC \
    --grid-lines=true \
    --header-font-size=12 \
    --header-bold=true
```

Or directly with the Go binary:

```bash
./gopdfconv --input=data.xlsx --output=report.pdf \
    --row-height=25 \
    --header-height=30 \
    --cell-padding=6 \
    --header-color=4A90D9 \
    --header-text-color=FFFFFF \
    --row-color=F5F5F5 \
    --border-color=CCCCCC
```

### Page Size Options

```php
PdfConverter::csv('data.csv')
    ->a4()        // 210 × 297 mm (default)
    ->a3()        // 297 × 420 mm
    ->letter()    // 8.5 × 11 inches
    ->legal()     // 8.5 × 14 inches
    ->tabloid()   // 11 × 17 inches
    ->convert();
```

### Storage Disk Integration

The `disk()` method allows using Laravel Storage disks to resolve file paths:

```php
// Using local Laravel Storage Disk
PdfConverter::csv('reports/data.csv')
    ->disk('local')
    ->toPdf('exports/data.pdf')
    ->convert();

// Using public disk
PdfConverter::csv('reports/data.csv')
    ->disk('public')
    ->toPdf('exports/data.pdf')
    ->convert();
```

### Cloud Storage (S3, GCS, Azure, etc.)

The package fully supports cloud storage. When using a cloud disk, files are automatically:

1. Downloaded from cloud to a local temp directory
2. Converted to PDF locally using the Go binary
3. Uploaded back to cloud storage
4. Temp files are cleaned up automatically

```php
// Amazon S3
PdfConverter::csv('reports/data.csv')
    ->disk('s3')
    ->toPdf('exports/data.pdf')
    ->convert();

// Google Cloud Storage
PdfConverter::excel('reports/sales.xlsx')
    ->disk('gcs')
    ->toPdf('exports/sales.pdf')
    ->convert();

// Azure Blob Storage
PdfConverter::pptx('presentations/deck.pptx')
    ->disk('azure')
    ->toPdf('exports/deck.pdf')
    ->convert();
```

**Supported Cloud Providers:**

- Amazon S3 (`s3`)
- Google Cloud Storage (`gcs`)
- Azure Blob Storage (`azure`)
- DigitalOcean Spaces (uses S3 driver)
- MinIO (uses S3 driver)
- FTP/SFTP (`ftp`, `sftp`)
- Dropbox (`dropbox`)

**Cloud Storage Configuration Example (config/filesystems.php):**

```php
'disks' => [
    's3' => [
        'driver' => 's3',
        'key' => env('AWS_ACCESS_KEY_ID'),
        'secret' => env('AWS_SECRET_ACCESS_KEY'),
        'region' => env('AWS_DEFAULT_REGION'),
        'bucket' => env('AWS_BUCKET'),
    ],

    'gcs' => [
        'driver' => 'gcs',
        'project_id' => env('GOOGLE_CLOUD_PROJECT_ID'),
        'key_file' => env('GOOGLE_CLOUD_KEY_FILE'),
        'bucket' => env('GOOGLE_CLOUD_STORAGE_BUCKET'),
    ],
],
```

**Batch Processing with Cloud Storage:**

```php
$result = PdfConverter::batch([
    'reports/q1.csv',
    'reports/q2.xlsx',
    'reports/q3.csv',
])
->disk('s3')
->outputDir('exports/pdfs')
->workers(4)
->convert();

// Result includes cloud paths
// $result['cloud_output_dir'] => 'exports/pdfs'
// $result['results'][0]['cloud_output_path'] => 'exports/pdfs/q1.pdf'
```

**Queue Jobs with Cloud Storage:**

```php
// Single file to queue with S3
PdfConverter::csv('reports/data.csv')
    ->disk('s3')
    ->toPdf('exports/data.pdf')
    ->queue();

// Batch to queue with cloud storage
PdfConverter::batch($files)
    ->disk('s3')
    ->outputDir('exports')
    ->queue();
```

**Important Notes:**

- Cloud operations require the appropriate Flysystem adapter package installed
- For S3: `composer require league/flysystem-aws-s3-v3`
- For GCS: `composer require league/flysystem-google-cloud-storage`
- Temp files are stored in `config('gopdf.temp_dir')` and cleaned up after conversion
- Large files may take longer due to download/upload time

### Batch Processing

Convert multiple files in parallel using Go's goroutines:

```php
$result = PdfConverter::batch([
    'file1.csv',
    'file2.xlsx',
])
->outputDir('/path/to/output')
->workers(4)    // Number of parallel workers (default: CPU cores, max: 16)
->convert();
```

**Verified Return Format:**

```php
[
    'total_jobs' => 2,
    'successful' => 2,
    'failed' => 0,
    'total_time_ns' => 36747200,
    'results' => [
        [
            'job' => ['ID' => 'job-1', 'InputPath' => '...', 'OutputPath' => '...'],
            'success' => true,
            'process_time_ns' => 33474500,
        ],
        // ...
    ]
]
```

### Queue Jobs

Dispatch conversions to Laravel queues:

```php
// Dispatch single file to queue
PdfConverter::csv('data.csv')
    ->toPdf('output.pdf')
    ->queue();

// With specific connection and queue
PdfConverter::csv('data.csv')
    ->toPdf('output.pdf')
    ->queue('redis', 'pdf-conversions');

// Batch to queue
PdfConverter::batch($files)
    ->outputDir($outputDir)
    ->queue();
```

**Queue Job Configuration (from config):**

- `tries`: 3 (retry attempts)
- `backoff`: 30 seconds between retries
- `timeout`: 120 seconds (single), 600 seconds (batch)
- Default queue name: `pdf-conversions`

**Note:** Queue functionality requires a properly configured Laravel queue driver (database, redis, etc.)

### Artisan Command

```bash
# Basic conversion
php artisan pdf:convert input.csv output.pdf

# With options
php artisan pdf:convert input.xlsx output.pdf \
    --page-size=Letter \
    --landscape \
    --header-text="My Report"   # Page header (centered at top)

# Async via queue
php artisan pdf:convert input.csv output.pdf --queue
```

---

## Supported Formats

| Format            | Extension        | Method               | Requirements | Table Styling |
| ----------------- | ---------------- | -------------------- | ------------ | ------------- |
| CSV               | `.csv`           | Native Go            | None         | ✅ Full       |
| TSV               | `.tsv`           | Native Go            | None         | ✅ Full       |
| Excel             | `.xlsx`, `.xlsm` | Native Go            | None         | ✅ Full       |
| Excel Legacy      | `.xls`           | LibreOffice → Native | LibreOffice  | ✅ Full       |
| PowerPoint        | `.pptx`          | LibreOffice          | LibreOffice  | ❌ Not supported |
| PowerPoint Legacy | `.ppt`           | LibreOffice          | LibreOffice  | ❌ Not supported |

> **Table Styling Column:** Indicates whether table customization options (colors, row heights, column widths, cell padding, font styling, grid lines) are supported. PowerPoint files use slide-based rendering and only support general options (page size, orientation, margins, watermark, header/footer).

### Conversion Details

- **CSV/TSV**: Parsed natively with auto-delimiter detection, rendered as professional tables
- **XLSX/XLSM**: Parsed natively using excelize library, supports multiple sheets
- **XLS**: Converted to XLSX via LibreOffice, then processed natively for table rendering
- **PPTX/PPT**: Converted via LibreOffice for full visual fidelity (backgrounds, images, layouts)

---

## Configuration

Publish the config file:

```bash
php artisan vendor:publish --tag=gopdf-config
```

```php
// config/gopdf.php

return [
    // Go binary path (auto-detected based on OS/architecture)
    'binary_path' => env('GOPDF_BINARY_PATH', null),

    // LibreOffice path (auto-detected if not set)
    'libreoffice_path' => env('GOPDF_LIBREOFFICE_PATH', null),

    // Temporary directory for conversions
    'temp_dir' => env('GOPDF_TEMP_DIR', storage_path('app/temp/gopdf')),

    // Default page settings
    'defaults' => [
        'page_size' => 'A4',        // A4, Letter, Legal, A3, Tabloid
        'orientation' => 'portrait', // portrait, landscape
        'margin' => 20,              // points
        'font_size' => 10,           // points
        'header_row' => true,        // Style first row as header
    ],

    // Timeout settings (seconds)
    'timeout' => [
        'single' => 120,    // Single file conversion
        'batch' => 600,     // Batch conversion
    ],

    // Queue settings
    'queue' => [
        'connection' => env('GOPDF_QUEUE_CONNECTION', null),
        'queue' => env('GOPDF_QUEUE_NAME', 'pdf-conversions'),
        'tries' => 3,       // Retry attempts
        'backoff' => 30,    // Seconds between retries
    ],

    // Batch processing
    'batch' => [
        'workers' => env('GOPDF_BATCH_WORKERS', 0), // 0 = auto (CPU cores)
    ],

    // Logging
    'logging' => [
        'enabled' => env('GOPDF_LOGGING', true),
        'channel' => env('GOPDF_LOG_CHANNEL', null),
    ],
];
```

### Environment Variables

```env
# Binary paths (usually auto-detected)
GOPDF_BINARY_PATH=
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

## Error Handling

```php
use NikunjKothiya\GoPdfConverter\Facades\PdfConverter;
use NikunjKothiya\GoPdfConverter\Exceptions\PdfConversionException;
use NikunjKothiya\GoPdfConverter\Exceptions\FileNotFoundException;
use NikunjKothiya\GoPdfConverter\Exceptions\UnsupportedFormatException;

try {
    PdfConverter::csv('data.csv')->toPdf('output.pdf')->convert();
} catch (FileNotFoundException $e) {
    echo "File not found: " . $e->getInputFile();
} catch (UnsupportedFormatException $e) {
    echo "Format not supported: " . $e->getMessage();
} catch (PdfConversionException $e) {
    echo "Error: " . $e->getMessage();
    echo "Code: " . $e->getErrorCode();
}
```

---

## Troubleshooting

### LibreOffice Not Found

For PPT, PPTX, and XLS files, LibreOffice must be installed:

```bash
# Check if LibreOffice is installed
libreoffice --version

# Set path in .env if not auto-detected
GOPDF_LIBREOFFICE_PATH=/usr/bin/libreoffice
```

### Permission Denied on Binary

```bash
chmod +x vendor/nikunjkothiya/laravel-go-pdf-converter/bin/gopdfconv-*
```

### Large File Timeout

```php
PdfConverter::csv('huge.csv')
    ->timeout(600)  // 10 minutes
    ->convert();
```

### PowerPoint Shows Only Text (No Images/Backgrounds)

This happens when LibreOffice is not available. Install LibreOffice for full fidelity conversion.

If you intentionally want text-only extraction:

```php
PdfConverter::pptx('presentation.pptx')
    ->native()  // Force native mode (text only)
    ->convert();
```

---

## Performance & Limitations

### Memory Optimization

The package is optimized for low memory consumption, even with very large files:

**Key Optimizations:**

| Optimization | Description |
|-------------|-------------|
| **Streaming Processing** | Files are processed row-by-row, never loading entire file into memory |
| **Two-Pass Streaming** | First pass samples 100 rows for column widths, second pass streams to PDF |
| **Memory Limits** | Excel decompression limited to 100MB (configurable) |
| **Buffered I/O** | 64KB read buffers for efficient disk access |
| **String Builder** | Uses `strings.Builder` for efficient text concatenation |
| **Pre-allocation** | Buffers pre-allocated to reduce garbage collection |

**Memory Usage Comparison:**

| File Size | Rows | Old Memory | New Memory (Streaming) |
|-----------|------|------------|------------------------|
| 10MB CSV | 50,000 | ~200MB | ~15MB |
| 50MB XLSX | 100,000 | ~500MB | ~60MB |
| 100MB XLSX | 500,000 | ~1GB+ | ~80MB |

### Large File Handling

The package processes **all rows** without artificial limits. It uses memory-efficient techniques:

- **Streaming Processing**: Files are processed row-by-row instead of loading everything into memory
- **Streaming Excel Reader**: Uses excelize streaming API (`Rows()`) with memory limits to handle large Excel files (100,000+ rows) efficiently
- **Streaming CSV Reader**: CSV files are streamed directly to PDF without loading all records
- **Buffered I/O**: Files are read with 64KB buffers for efficient disk access
- **Column Width Sampling**: Only first 100 rows are sampled for column width calculation
- **Optimized String Handling**: Uses `strings.Builder` for efficient text concatenation
- **Pre-allocated Buffers**: Memory is pre-allocated where possible to reduce GC pressure
- **Memory Limits**: Excel files have configurable unzip size limits (100MB default)

**Memory Usage Tips:**

For very large files (100,000+ rows), the streaming approach keeps memory usage low:
- CSV: ~10-20MB for any file size
- Excel: ~50-100MB depending on file complexity

```php
// Large file conversion - memory efficient by default
PdfConverter::excel('huge_file.xlsx')
    ->timeout(600)  // 10 minutes for large files
    ->convert();
```

**Timeout Configuration:**

```php
// Increase timeout for very large files
PdfConverter::csv('large_file.csv')
    ->timeout(600)  // 10 minutes
    ->convert();
```

### Performance Notes

- **CSV/TSV**: Native Go parsing with buffered I/O, handles large files efficiently
- **XLSX/XLSM**: Native Go parsing using excelize streaming API, processes all sheets and rows without row limits
- **XLS**: Requires LibreOffice to convert to XLSX first, then native processing
- **PPTX/PPT**: Requires LibreOffice for full visual fidelity

**Factors affecting performance:**

- File size and number of rows/columns
- Number of sheets (Excel)
- Complexity of content (PowerPoint)
- System resources (CPU, memory, disk speed)
- LibreOffice startup time (for formats requiring it)

---

## Building from Source

If you need to rebuild the Go binary:

```bash
cd vendor/nikunjkothiya/laravel-go-pdf-converter/go-binary
go mod download
go build -o ../bin/gopdfconv-linux-amd64 ./cmd/gopdfconv      # Linux
go build -o ../bin/gopdfconv-darwin-amd64 ./cmd/gopdfconv     # macOS
go build -o ../bin/gopdfconv-windows-amd64.exe ./cmd/gopdfconv # Windows
```

---

## License

This package is open-source software licensed under the **MIT license**.

### Third-Party Dependencies

| Library                                                       | License      | Purpose             |
| ------------------------------------------------------------- | ------------ | ------------------- |
| [signintech/gopdf](https://github.com/signintech/gopdf)       | MIT          | PDF generation      |
| [xuri/excelize](https://github.com/xuri/excelize)             | BSD 3-Clause | Excel parsing       |
| [richardlehane/mscfb](https://github.com/richardlehane/mscfb) | Apache 2.0   | Legacy file parsing |

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
