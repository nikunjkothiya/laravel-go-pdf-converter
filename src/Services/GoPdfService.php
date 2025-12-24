<?php

namespace NikunjKothiya\GoPdfConverter\Services;

use Illuminate\Support\Facades\Process;
use Illuminate\Support\Facades\Log;
use NikunjKothiya\GoPdfConverter\Exceptions\PdfConversionException;
use NikunjKothiya\GoPdfConverter\Exceptions\FileNotFoundException;
use NikunjKothiya\GoPdfConverter\Exceptions\UnsupportedFormatException;
use NikunjKothiya\GoPdfConverter\Exceptions\BinaryNotFoundException;
use NikunjKothiya\GoPdfConverter\Exceptions\TimeoutException;
use NikunjKothiya\GoPdfConverter\PdfBuilder;
use NikunjKothiya\GoPdfConverter\BatchBuilder;

class GoPdfService
{
    protected ?string $binaryPath;
    protected ?string $libreOfficePath;
    protected string $tempDir;
    protected array $defaults;
    protected array $timeouts;

    protected const SUPPORTED_FORMATS = ['csv', 'tsv', 'xlsx', 'xls', 'xlsm', 'pptx', 'ppt'];

    public function __construct(
        ?string $binaryPath = null,
        ?string $libreOfficePath = null,
        ?string $tempDir = null,
        array $defaults = [],
        array $timeouts = []
    ) {
        $this->binaryPath = $binaryPath;
        $this->libreOfficePath = $libreOfficePath;
        $this->tempDir = $tempDir ?? sys_get_temp_dir();
        $this->defaults = array_merge([
            'page_size' => 'A4',
            'orientation' => 'portrait',
            'margin' => 20,
            'font_size' => 10,
            'header_row' => true,
        ], $defaults);
        $this->timeouts = array_merge([
            'single' => 300,
            'batch' => 600,
        ], $timeouts);
    }

    /**
     * Create a builder for CSV conversion
     */
    public function csv(string $inputPath): PdfBuilder
    {
        return $this->from($inputPath)->format('csv');
    }

    /**
     * Create a builder for Excel conversion
     */
    public function excel(string $inputPath): PdfBuilder
    {
        return $this->from($inputPath)->format('xlsx');
    }

    /**
     * Alias for excel()
     */
    public function xlsx(string $inputPath): PdfBuilder
    {
        return $this->excel($inputPath);
    }

    /**
     * Create a builder for PowerPoint conversion
     */
    public function pptx(string $inputPath): PdfBuilder
    {
        return $this->from($inputPath)->format('pptx');
    }

    /**
     * Alias for pptx()
     */
    public function powerpoint(string $inputPath): PdfBuilder
    {
        return $this->pptx($inputPath);
    }

    /**
     * Create a builder from any supported file
     */
    public function from(string $inputPath): PdfBuilder
    {
        return new PdfBuilder($this, $inputPath);
    }

    /**
     * Create a batch builder
     */
    public function batch(array $inputPaths): BatchBuilder
    {
        return new BatchBuilder($this, $inputPaths);
    }

    /**
     * Perform the actual conversion
     * 
     * @throws PdfConversionException
     */
    public function convert(string $inputPath, string $outputPath, array $options = []): array
    {
        // Validate input file
        if (!file_exists($inputPath)) {
            throw new FileNotFoundException($inputPath);
        }

        // Validate format
        $extension = strtolower(pathinfo($inputPath, PATHINFO_EXTENSION));
        if (!in_array($extension, self::SUPPORTED_FORMATS)) {
            throw new UnsupportedFormatException($inputPath, $extension);
        }

        // Get binary path
        $binary = $this->resolveBinaryPath();
        if (!$binary || !file_exists($binary)) {
            throw new BinaryNotFoundException($binary);
        }

        // Ensure output directory exists
        $outputDir = dirname($outputPath);
        if (!is_dir($outputDir)) {
            mkdir($outputDir, 0755, true);
        }

        // Merge options with defaults
        $options = array_merge($this->defaults, $options);

        // Build command
        $command = $this->buildCommand($binary, $inputPath, $outputPath, $options);

        // Log if enabled
        if (config('gopdf.logging.enabled', true)) {
            Log::channel(config('gopdf.logging.channel'))
                ->debug('GoPdf: Starting conversion', [
                    'input' => $inputPath,
                    'output' => $outputPath,
                    'options' => $options,
                ]);
        }

        // Execute conversion
        $timeout = $options['timeout'] ?? $this->timeouts['single'];
        
        $result = Process::timeout($timeout)
            ->run($command);

        // Parse result
        $output = $result->output();
        $exitCode = $result->exitCode();

        // Try to parse JSON output
        $data = json_decode($output, true);

        if ($result->failed() || ($data && !($data['success'] ?? false))) {
            $this->handleError($data, $inputPath, $result->errorOutput());
        }

        // Log success
        if (config('gopdf.logging.enabled', true)) {
            Log::channel(config('gopdf.logging.channel'))
                ->info('GoPdf: Conversion completed', [
                    'input' => $inputPath,
                    'output' => $outputPath,
                    'time_ms' => $data['process_time_ms'] ?? null,
                    'size' => $data['file_size_bytes'] ?? null,
                ]);
        }

        return [
            'success' => true,
            'input_file' => $inputPath,
            'output_file' => $outputPath,
            'format' => $data['format'] ?? $extension,
            'process_time_ms' => $data['process_time_ms'] ?? null,
            'file_size_bytes' => $data['file_size_bytes'] ?? filesize($outputPath),
        ];
    }

    /**
     * Perform batch conversion
     */
    public function convertBatch(array $files, string $outputDir, array $options = []): array
    {
        // Get binary path
        $binary = $this->resolveBinaryPath();
        if (!$binary || !file_exists($binary)) {
            throw new BinaryNotFoundException($binary);
        }

        // Ensure output directory exists
        if (!is_dir($outputDir)) {
            mkdir($outputDir, 0755, true);
        }

        // Build batch input
        $inputFiles = implode(',', array_map('escapeshellarg', $files));

        // Build command
        $options = array_merge($this->defaults, $options);
        $command = [
            $binary,
            '--batch=' . implode(',', $files),
            '--output-dir=' . $outputDir,
            '--json',
        ];

        $this->addOptionsToCommand($command, $options);

        // Execute
        $timeout = $options['timeout'] ?? $this->timeouts['batch'];
        
        $result = Process::timeout($timeout)
            ->run($command);

        $output = $result->output();
        $data = json_decode($output, true);

        if (!$data) {
            throw new PdfConversionException(
                'Batch conversion failed: Invalid response',
                'CONVERSION_FAILED',
                null,
                $result->errorOutput()
            );
        }

        return $data;
    }

    /**
     * Check if the binary is available
     */
    public function isAvailable(): bool
    {
        $binary = $this->resolveBinaryPath();
        return $binary && file_exists($binary) && is_executable($binary);
    }

    /**
     * Get the resolved binary path
     */
    public function getBinaryPath(): string
    {
        return $this->resolveBinaryPath() ?? '';
    }

    /**
     * Get supported formats
     */
    public function getSupportedFormats(): array
    {
        return self::SUPPORTED_FORMATS;
    }

    /**
     * Resolve the binary path based on OS/architecture
     */
    protected function resolveBinaryPath(): ?string
    {
        // Use configured path if set
        if ($this->binaryPath && file_exists($this->binaryPath)) {
            return $this->binaryPath;
        }

        // Auto-detect based on OS/arch
        $os = PHP_OS_FAMILY === 'Windows' ? 'windows' : strtolower(PHP_OS_FAMILY);
        $arch = php_uname('m');

        // Normalize architecture
        if (in_array($arch, ['x86_64', 'amd64', 'AMD64'])) {
            $arch = 'amd64';
        } elseif (in_array($arch, ['aarch64', 'arm64', 'ARM64'])) {
            $arch = 'arm64';
        }

        // Normalize OS
        if ($os === 'darwin') {
            $os = 'darwin';
        } elseif ($os === 'linux') {
            $os = 'linux';
        } elseif ($os === 'windows') {
            $os = 'windows';
        }

        // Build binary name
        $binaryName = "gopdfconv-{$os}-{$arch}";
        if ($os === 'windows') {
            $binaryName .= '.exe';
        }

        // Check in package bin directory
        $packageBinPath = dirname(__DIR__, 2) . '/bin/' . $binaryName;
        if (file_exists($packageBinPath)) {
            return $packageBinPath;
        }

        // Check in vendor bin
        $vendorBinPath = base_path('vendor/bin/' . $binaryName);
        if (file_exists($vendorBinPath)) {
            return $vendorBinPath;
        }

        // Check system path
        $systemPaths = [
            '/usr/local/bin/gopdfconv',
            '/usr/bin/gopdfconv',
        ];

        foreach ($systemPaths as $path) {
            if (file_exists($path)) {
                return $path;
            }
        }

        return null;
    }

    /**
     * Build the conversion command
     */
    protected function buildCommand(string $binary, string $input, string $output, array $options): array
    {
        $command = [
            $binary,
            '--input=' . $input,
            '--output=' . $output,
            '--json',
        ];

        $this->addOptionsToCommand($command, $options);

        return $command;
    }

    /**
     * Add options to command array
     */
    protected function addOptionsToCommand(array &$command, array $options): void
    {
        if (isset($options['format']) && $options['format'] !== 'auto') {
            $command[] = '--format=' . $options['format'];
        }

        if (isset($options['page_size'])) {
            $command[] = '--page-size=' . $options['page_size'];
        }

        if (isset($options['orientation'])) {
            $command[] = '--orientation=' . $options['orientation'];
        }

        if (isset($options['margin'])) {
            $command[] = '--margin=' . $options['margin'];
        }

        if (isset($options['font_size'])) {
            $command[] = '--font-size=' . $options['font_size'];
        }

        if (isset($options['header_row'])) {
            $command[] = '--header=' . ($options['header_row'] ? 'true' : 'false');
        }

        if (isset($options['workers'])) {
            $command[] = '--workers=' . $options['workers'];
        }

        if (isset($options['native']) && $options['native']) {
            $command[] = '--native';
        }

        if (isset($options['header_text']) && $options['header_text']) {
            $command[] = '--header-text=' . $options['header_text'];
        }

        if (isset($options['footer_text']) && $options['footer_text']) {
            $command[] = '--footer-text=' . $options['footer_text'];
        }

        if ($this->libreOfficePath) {
            $command[] = '--libreoffice=' . $this->libreOfficePath;
        }

        // Advanced features
        if (isset($options['font'])) {
            $command[] = '--font=' . $options['font'];
        }
        if (isset($options['watermark_text'])) {
            $command[] = '--watermark-text=' . $options['watermark_text'];
        }
        if (isset($options['watermark_image'])) {
            $command[] = '--watermark-image=' . $options['watermark_image'];
        }
        if (isset($options['watermark_alpha'])) {
            $command[] = '--watermark-alpha=' . $options['watermark_alpha'];
        }
        if (isset($options['header_color'])) {
            $command[] = '--header-color=' . $options['header_color'];
        }
        if (isset($options['row_color'])) {
            $command[] = '--row-color=' . $options['row_color'];
        }
        if (isset($options['border_color'])) {
            $command[] = '--border-color=' . $options['border_color'];
        }
        if (isset($options['grid_lines'])) {
            $command[] = '--grid-lines=' . ($options['grid_lines'] ? 'true' : 'false');
        }
    }

    /**
     * Handle conversion error
     * 
     * @throws PdfConversionException
     */
    protected function handleError(?array $data, string $inputPath, string $stderr): void
    {
        if ($data && isset($data['error'])) {
            throw PdfConversionException::fromJson($data);
        }

        throw new PdfConversionException(
            'Conversion failed',
            'CONVERSION_FAILED',
            $inputPath,
            $stderr
        );
    }
}
