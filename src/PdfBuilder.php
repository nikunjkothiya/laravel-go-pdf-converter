<?php

namespace NikunjKothiya\GoPdfConverter;

use NikunjKothiya\GoPdfConverter\Services\GoPdfService;
use NikunjKothiya\GoPdfConverter\Jobs\ConvertToPdfJob;
use Illuminate\Support\Facades\Storage;

/**
 * Fluent builder for single file PDF conversion
 */
class PdfBuilder
{
    protected GoPdfService $service;
    protected string $inputPath;
    protected ?string $outputPath = null;
    protected ?string $disk = null;
    protected array $options = [];
    protected array $tempFiles = [];

    public function __construct(GoPdfService $service, string $inputPath)
    {
        $this->service = $service;
        $this->inputPath = $inputPath;
    }

    /**
     * Set the output file path
     */
    public function toPdf(string $outputPath): self
    {
        $this->outputPath = $outputPath;
        return $this;
    }

    /**
     * Set the storage disk to use for input and output paths
     */
    public function disk(string $disk): self
    {
        $this->disk = $disk;
        return $this;
    }

    /**
     * Alias for toPdf()
     */
    public function saveTo(string $outputPath): self
    {
        return $this->toPdf($outputPath);
    }

    /**
     * Force a specific input format
     */
    public function format(string $format): self
    {
        $this->options['format'] = strtolower($format);
        return $this;
    }

    /**
     * Set page size (A4, Letter, Legal, A3)
     */
    public function pageSize(string $size): self
    {
        $this->options['page_size'] = $size;
        return $this;
    }

    /**
     * Shorthand for A4 page size (210 × 297 mm)
     */
    public function a4(): self
    {
        return $this->pageSize('A4');
    }

    /**
     * Shorthand for A3 page size (297 × 420 mm) - good for wide tables
     */
    public function a3(): self
    {
        return $this->pageSize('A3');
    }

    /**
     * Shorthand for Letter page size (8.5 × 11 inches)
     */
    public function letter(): self
    {
        return $this->pageSize('Letter');
    }

    /**
     * Shorthand for Legal page size (8.5 × 14 inches)
     */
    public function legal(): self
    {
        return $this->pageSize('Legal');
    }

    /**
     * Shorthand for Tabloid page size (11 × 17 inches) - best for very wide tables
     */
    public function tabloid(): self
    {
        return $this->pageSize('Tabloid');
    }

    /**
     * Auto landscape mode - recommended for CSVs with many columns
     * Uses A3 landscape for optimal viewing
     */
    public function wideFormat(): self
    {
        return $this->pageSize('A3')->landscape()->fontSize(8)->margin(10);
    }

    /**
     * Set page orientation
     */
    public function orientation(string $orientation): self
    {
        $this->options['orientation'] = strtolower($orientation);
        return $this;
    }

    /**
     * Set portrait orientation
     */
    public function portrait(): self
    {
        return $this->orientation('portrait');
    }

    /**
     * Set landscape orientation
     */
    public function landscape(): self
    {
        return $this->orientation('landscape');
    }

    /**
     * Set page margins (in points)
     */
    public function margin(float $margin): self
    {
        $this->options['margin'] = $margin;
        return $this;
    }

    /**
     * Set font size
     */
    public function fontSize(float $size): self
    {
        $this->options['font_size'] = $size;
        return $this;
    }

    /**
     * Enable header row styling (CSV/Excel)
     */
    public function withHeaders(): self
    {
        $this->options['header_row'] = true;
        return $this;
    }

    /**
     * Disable header row styling (CSV/Excel)
     */
    public function withoutHeaders(): self
    {
        $this->options['header_row'] = false;
        return $this;
    }

    /**
     * Set conversion timeout in seconds
     */
    public function timeout(int $seconds): self
    {
        $this->options['timeout'] = $seconds;
        return $this;
    }

    /**
     * Force native Go conversion (bypass LibreOffice)
     */
    public function native(bool $native = true): self
    {
        $this->options['native'] = $native;
        return $this;
    }

    /**
     * Set global header text
     */
    public function headerText(string $text): self
    {
        $this->options['header_text'] = $text;
        return $this;
    }

    /**
     * Set global footer text (Left aligned)
     */
    public function footerText(string $text): self
    {
        $this->options['footer_text'] = $text;
        return $this;
    }

    /**
     * Enable/Disable auto-orientation (Smart Layout)
     */
    public function autoOrientation(bool $enable = true): self
    {
        $this->options['auto_orientation'] = $enable;
        return $this;
    }

    /**
     * Set custom TTF font path
     */
    public function font(string $path): self
    {
        $this->options['font'] = $path;
        return $this;
    }

    /**
     * Add text watermark
     */
    public function watermarkText(string $text, float $alpha = 0.2): self
    {
        $this->options['watermark_text'] = $text;
        $this->options['watermark_alpha'] = $alpha;
        return $this;
    }

    /**
     * Add image watermark
     */
    public function watermarkImage(string $path, float $alpha = 0.2): self
    {
        $this->options['watermark_image'] = $path;
        $this->options['watermark_alpha'] = $alpha;
        return $this;
    }

    /**
     * Set table header background color (hex)
     */
    public function headerColor(string $hex): self
    {
        $this->options['header_color'] = $hex;
        return $this;
    }

    /**
     * Set alternating row color (hex)
     */
    public function rowColor(string $hex): self
    {
        $this->options['row_color'] = $hex;
        return $this;
    }

    /**
     * Set table border color (hex)
     */
    public function borderColor(string $hex): self
    {
        $this->options['border_color'] = $hex;
        return $this;
    }

    /**
     * Toggle grid lines
     */
    public function showGridLines(bool $show = true): self
    {
        $this->options['grid_lines'] = $show;
        return $this;
    }

    /**
     * Set custom row height (0 = auto)
     */
    public function rowHeight(float $height): self
    {
        $this->options['row_height'] = $height;
        return $this;
    }

    /**
     * Set custom header row height (0 = auto)
     */
    public function headerHeight(float $height): self
    {
        $this->options['header_height'] = $height;
        return $this;
    }

    /**
     * Set cell padding in points
     */
    public function cellPadding(float $padding): self
    {
        $this->options['cell_padding'] = $padding;
        return $this;
    }

    /**
     * Set minimum column width
     */
    public function minColumnWidth(float $width): self
    {
        $this->options['min_col_width'] = $width;
        return $this;
    }

    /**
     * Set maximum column width
     */
    public function maxColumnWidth(float $width): self
    {
        $this->options['max_col_width'] = $width;
        return $this;
    }

    /**
     * Set header text color (hex)
     */
    public function headerTextColor(string $hex): self
    {
        $this->options['header_text_color'] = $hex;
        return $this;
    }

    /**
     * Set row text color (hex)
     */
    public function rowTextColor(string $hex): self
    {
        $this->options['row_text_color'] = $hex;
        return $this;
    }

    /**
     * Set header font size (0 = auto)
     */
    public function headerFontSize(float $size): self
    {
        $this->options['header_font_size'] = $size;
        return $this;
    }

    /**
     * Set header text bold
     */
    public function headerBold(bool $bold = true): self
    {
        $this->options['header_bold'] = $bold;
        return $this;
    }

    /**
     * Add custom option
     */
    public function option(string $key, mixed $value): self
    {
        $this->options[$key] = $value;
        return $this;
    }

    /**
     * Add multiple options at once
     */
    public function options(array $options): self
    {
        $this->options = array_merge($this->options, $options);
        return $this;
    }

    /**
     * Execute the conversion synchronously
     * 
     * @return array Conversion result
     */
    public function convert(): array
    {
        $isCloud = $this->disk && $this->isCloudDisk($this->disk);
        
        if ($isCloud) {
            return $this->convertFromCloud();
        }
        
        $inputPath = $this->resolvePath($this->inputPath);
        $outputPath = $this->resolvePath($this->outputPath ?? $this->generateOutputPath());
        
        return $this->service->convert($inputPath, $outputPath, $this->options);
    }

    /**
     * Handle conversion for cloud storage (S3, GCS, etc.)
     * Downloads file from cloud, converts locally, uploads result back
     */
    protected function convertFromCloud(): array
    {
        try {
            // Download input file from cloud to temp location
            $localInputPath = $this->downloadFromCloud($this->inputPath);
            
            // Generate local temp output path
            $outputFileName = $this->outputPath ?? $this->generateOutputPath();
            $localOutputPath = $this->getTempPath(basename($outputFileName));
            
            // Perform conversion locally
            $result = $this->service->convert($localInputPath, $localOutputPath, $this->options);
            
            if ($result['success'] && file_exists($localOutputPath)) {
                // Upload converted PDF back to cloud
                $this->uploadToCloud($localOutputPath, $outputFileName);
                
                // Update result with cloud path
                $result['output_file'] = $outputFileName;
                $result['storage_disk'] = $this->disk;
                $result['cloud_storage'] = true;
            }
            
            return $result;
        } finally {
            // Cleanup temp files
            $this->cleanupTempFiles();
        }
    }

    /**
     * Check if a disk is cloud-based (non-local)
     */
    protected function isCloudDisk(string $diskName): bool
    {
        try {
            $disk = Storage::disk($diskName);
            $adapter = $disk->getAdapter();
            
            // Check for common cloud adapter types
            $adapterClass = get_class($adapter);
            
            $cloudAdapters = [
                'League\Flysystem\AwsS3V3\AwsS3V3Adapter',
                'League\Flysystem\AzureBlobStorage\AzureBlobStorageAdapter',
                'League\Flysystem\GoogleCloudStorage\GoogleCloudStorageAdapter',
                'League\Flysystem\Ftp\FtpAdapter',
                'League\Flysystem\Sftp\SftpAdapter',
                'Spatie\FlysystemDropbox\DropboxAdapter',
            ];
            
            foreach ($cloudAdapters as $cloudAdapter) {
                if ($adapter instanceof $cloudAdapter || $adapterClass === $cloudAdapter) {
                    return true;
                }
            }
            
            // Also check if the disk config has a driver that's cloud-based
            $config = config("filesystems.disks.{$diskName}");
            if ($config) {
                $cloudDrivers = ['s3', 'gcs', 'azure', 'ftp', 'sftp', 'dropbox', 'rackspace'];
                if (isset($config['driver']) && in_array($config['driver'], $cloudDrivers)) {
                    return true;
                }
            }
            
            return false;
        } catch (\Exception $e) {
            // If we can't determine, assume local
            return false;
        }
    }

    /**
     * Download file from cloud storage to local temp directory
     */
    protected function downloadFromCloud(string $cloudPath): string
    {
        $disk = Storage::disk($this->disk);
        
        if (!$disk->exists($cloudPath)) {
            throw new \NikunjKothiya\GoPdfConverter\Exceptions\FileNotFoundException(
                "File not found on cloud storage: {$cloudPath}",
                $cloudPath
            );
        }
        
        $localPath = $this->getTempPath(basename($cloudPath));
        $contents = $disk->get($cloudPath);
        
        file_put_contents($localPath, $contents);
        $this->tempFiles[] = $localPath;
        
        return $localPath;
    }

    /**
     * Upload file to cloud storage
     */
    protected function uploadToCloud(string $localPath, string $cloudPath): bool
    {
        $disk = Storage::disk($this->disk);
        $contents = file_get_contents($localPath);
        
        return $disk->put($cloudPath, $contents);
    }

    /**
     * Get temp directory path for cloud operations
     */
    protected function getTempPath(string $filename): string
    {
        $tempDir = config('gopdf.temp_dir', sys_get_temp_dir() . '/gopdf');
        
        if (!is_dir($tempDir)) {
            mkdir($tempDir, 0755, true);
        }
        
        return $tempDir . '/' . uniqid('gopdf_') . '_' . $filename;
    }

    /**
     * Cleanup temporary files created during cloud conversion
     */
    protected function cleanupTempFiles(): void
    {
        foreach ($this->tempFiles as $file) {
            if (file_exists($file)) {
                @unlink($file);
            }
        }
        $this->tempFiles = [];
    }

    /**
     * Dispatch conversion to queue
     * 
     * @return \Illuminate\Foundation\Bus\PendingDispatch
     */
    public function queue(?string $connection = null, ?string $queue = null)
    {
        $isCloud = $this->disk && $this->isCloudDisk($this->disk);
        
        if ($isCloud) {
            // For cloud storage, pass disk info to job
            $inputPath = $this->inputPath;
            $outputPath = $this->outputPath ?? $this->generateOutputPath();
        } else {
            $inputPath = $this->resolvePath($this->inputPath);
            $outputPath = $this->resolvePath($this->outputPath ?? $this->generateOutputPath());
        }
        
        $options = $this->options;
        if ($isCloud) {
            $options['_cloud_disk'] = $this->disk;
            $options['_cloud_input'] = $this->inputPath;
            $options['_cloud_output'] = $this->outputPath ?? $this->generateOutputPath();
        }
        
        $job = new ConvertToPdfJob(
            $inputPath,
            $outputPath,
            $options
        );

        if ($connection) {
            $job->onConnection($connection);
        }

        if ($queue) {
            $job->onQueue($queue);
        }

        return dispatch($job);
    }

    /**
     * Dispatch conversion and wait for result
     */
    public function dispatchSync(): array
    {
        $outputPath = $this->outputPath ?? $this->generateOutputPath();
        
        $job = new ConvertToPdfJob(
            $this->inputPath,
            $outputPath,
            $this->options
        );

        dispatch_sync($job);

        return [
            'success' => true,
            'input_file' => $this->inputPath,
            'output_file' => $outputPath,
        ];
    }

    /**
     * Generate output path based on input
     */
    protected function generateOutputPath(): string
    {
        $info = pathinfo($this->inputPath);
        $dir = isset($info['dirname']) && $info['dirname'] !== '.' ? $info['dirname'] . '/' : '';
        return $dir . $info['filename'] . '.pdf';
    }

    /**
     * Resolve path using Laravel storage if disk is set
     */
    protected function resolvePath(string $path): string
    {
        if ($this->disk) {
            // For cloud disks, don't try to get local path
            if ($this->isCloudDisk($this->disk)) {
                return $path;
            }
            return Storage::disk($this->disk)->path($path);
        }

        // If path is absolute, return it
        if (str_starts_with($path, '/') || (strtoupper(substr(PHP_OS, 0, 3)) === 'WIN' && strpos($path, ':') === 1)) {
            return $path;
        }

        // Otherwise, assume it's relative to base_path()
        return base_path($path);
    }

    /**
     * Get the input path
     */
    public function getInputPath(): string
    {
        return $this->inputPath;
    }

    /**
     * Get the output path
     */
    public function getOutputPath(): ?string
    {
        return $this->outputPath;
    }

    /**
     * Get the options
     */
    public function getOptions(): array
    {
        return $this->options;
    }
}
