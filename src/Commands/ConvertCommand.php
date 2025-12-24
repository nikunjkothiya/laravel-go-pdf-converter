<?php

namespace NikunjKothiya\GoPdfConverter\Commands;

use Illuminate\Console\Command;
use NikunjKothiya\GoPdfConverter\Facades\PdfConverter;
use NikunjKothiya\GoPdfConverter\Exceptions\PdfConversionException;

class ConvertCommand extends Command
{
    /**
     * The name and signature of the console command.
     */
    protected $signature = 'pdf:convert
                            {input : Input file path (CSV, XLSX, or PPTX)}
                            {output? : Output PDF file path (optional, auto-generated if not provided)}
                            {--format= : Force input format (csv|xlsx|pptx)}
                            {--page-size=A4 : Page size (A4|Letter|Legal|A3)}
                            {--landscape : Use landscape orientation}
                            {--margin=20 : Page margin in points}
                            {--font-size=10 : Base font size}
                            {--no-headers : Disable header row styling for CSV/Excel}
                            {--timeout=120 : Conversion timeout in seconds}
                            {--queue : Process via queue (async)}
                            {--queue-connection= : Queue connection to use}
                            {--queue-name= : Queue name to use}
                            {--native : Force native Go conversion (bypass LibreOffice)}
                            {--header-text= : Global header text}
                            {--footer-text= : Global footer text}';

    /**
     * The console command description.
     */
    protected $description = 'Convert CSV, Excel, or PowerPoint files to PDF';

    /**
     * Execute the console command.
     */
    public function handle(): int
    {
        $inputPath = $this->argument('input');
        $outputPath = $this->argument('output');

        // Validate input file
        if (!file_exists($inputPath)) {
            $this->error("Input file not found: {$inputPath}");
            return Command::FAILURE;
        }

        // Generate output path if not provided
        if (!$outputPath) {
            $info = pathinfo($inputPath);
            $outputPath = $info['dirname'] . '/' . $info['filename'] . '.pdf';
        }

        // Build options
        $options = $this->buildOptions();

        // Show info
        $this->info("Converting: {$inputPath}");
        $this->info("Output: {$outputPath}");

        // Check if queued
        if ($this->option('queue')) {
            return $this->handleQueued($inputPath, $outputPath, $options);
        }

        // Synchronous conversion
        return $this->handleSync($inputPath, $outputPath, $options);
    }

    /**
     * Handle synchronous conversion
     */
    protected function handleSync(string $inputPath, string $outputPath, array $options): int
    {
        $this->info('Starting conversion...');
        
        $progressBar = $this->output->createProgressBar(100);
        $progressBar->start();
        $progressBar->advance(10);

        try {
            $result = PdfConverter::convert($inputPath, $outputPath, $options);
            
            $progressBar->finish();
            $this->newLine();

            $this->info('✓ Conversion completed successfully!');
            $this->table(
                ['Property', 'Value'],
                [
                    ['Input File', $result['input_file']],
                    ['Output File', $result['output_file']],
                    ['Format', $result['format'] ?? 'auto'],
                    ['Process Time', ($result['process_time_ms'] ?? 0) . ' ms'],
                    ['File Size', $this->formatBytes($result['file_size_bytes'] ?? 0)],
                ]
            );

            return Command::SUCCESS;

        } catch (PdfConversionException $e) {
            $progressBar->finish();
            $this->newLine();
            
            $this->error('Conversion failed: ' . $e->getMessage());
            
            if ($e->getDetails()) {
                $this->warn('Details: ' . $e->getDetails());
            }

            return Command::FAILURE;

        } catch (\Exception $e) {
            $progressBar->finish();
            $this->newLine();
            
            $this->error('Unexpected error: ' . $e->getMessage());
            
            return Command::FAILURE;
        }
    }

    /**
     * Handle queued conversion
     */
    protected function handleQueued(string $inputPath, string $outputPath, array $options): int
    {
        try {
            $builder = PdfConverter::from($inputPath)
                ->toPdf($outputPath)
                ->options($options);

            $connection = $this->option('queue-connection');
            $queue = $this->option('queue-name');

            $builder->queue($connection, $queue);

            $this->info('✓ Conversion job dispatched to queue');
            $this->info("  Queue: " . ($queue ?? config('gopdf.queue.queue', 'pdf-conversions')));
            $this->info("  Connection: " . ($connection ?? 'default'));

            return Command::SUCCESS;

        } catch (\Exception $e) {
            $this->error('Failed to dispatch job: ' . $e->getMessage());
            return Command::FAILURE;
        }
    }

    /**
     * Build options array from command options
     */
    protected function buildOptions(): array
    {
        $options = [];

        if ($format = $this->option('format')) {
            $options['format'] = $format;
        }

        $options['page_size'] = $this->option('page-size');
        $options['orientation'] = $this->option('landscape') ? 'landscape' : 'portrait';
        $options['margin'] = (float) $this->option('margin');
        $options['font_size'] = (float) $this->option('font-size');
        $options['header_row'] = !$this->option('no-headers');
        $options['timeout'] = (int) $this->option('timeout');
        $options['native'] = $this->option('native');
        $options['header_text'] = $this->option('header-text');
        $options['footer_text'] = $this->option('footer-text');

        return $options;
    }

    /**
     * Format bytes to human readable
     */
    protected function formatBytes(int $bytes): string
    {
        $units = ['B', 'KB', 'MB', 'GB'];
        $factor = floor((strlen((string) $bytes) - 1) / 3);
        return sprintf("%.2f %s", $bytes / pow(1024, $factor), $units[$factor]);
    }
}
