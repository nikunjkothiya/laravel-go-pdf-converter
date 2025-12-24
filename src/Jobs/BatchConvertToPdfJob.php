<?php

namespace NikunjKothiya\GoPdfConverter\Jobs;

use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;
use NikunjKothiya\GoPdfConverter\Facades\PdfConverter;
use NikunjKothiya\GoPdfConverter\Exceptions\PdfConversionException;

class BatchConvertToPdfJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public int $tries;
    public int $backoff;
    public int $timeout;

    protected array $inputPaths;
    protected string $outputDir;
    protected array $options;

    /**
     * Create a new job instance.
     */
    public function __construct(
        array $inputPaths,
        string $outputDir,
        array $options = []
    ) {
        $this->inputPaths = $inputPaths;
        $this->outputDir = $outputDir;
        $this->options = $options;

        $this->tries = config('gopdf.queue.tries', 3);
        $this->backoff = config('gopdf.queue.backoff', 30);
        $this->timeout = config('gopdf.timeout.batch', 600);

        $this->onQueue(config('gopdf.queue.queue', 'pdf-conversions'));
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        Log::info('GoPdf Batch Job: Starting batch conversion', [
            'file_count' => count($this->inputPaths),
            'output_dir' => $this->outputDir,
        ]);

        try {
            $result = app('gopdf.converter')->convertBatch(
                $this->inputPaths,
                $this->outputDir,
                $this->options
            );

            Log::info('GoPdf Batch Job: Batch conversion completed', [
                'total' => $result['total_jobs'] ?? count($this->inputPaths),
                'successful' => $result['successful'] ?? 0,
                'failed' => $result['failed'] ?? 0,
            ]);

            // If any failed, throw exception to trigger retry
            if (($result['failed'] ?? 0) > 0 && ($result['successful'] ?? 0) === 0) {
                throw new PdfConversionException(
                    "Batch conversion failed: all files failed",
                    'BATCH_FAILED'
                );
            }

        } catch (PdfConversionException $e) {
            Log::error('GoPdf Batch Job: Batch conversion failed', [
                'error' => $e->getMessage(),
            ]);

            throw $e;
        }
    }

    /**
     * Handle job failure.
     */
    public function failed(\Throwable $exception): void
    {
        Log::error('GoPdf Batch Job: Job failed permanently', [
            'file_count' => count($this->inputPaths),
            'output_dir' => $this->outputDir,
            'error' => $exception->getMessage(),
        ]);
    }

    /**
     * Get the tags for the job.
     */
    public function tags(): array
    {
        return [
            'gopdf',
            'batch',
            'count:' . count($this->inputPaths),
        ];
    }
}
