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

class ConvertToPdfJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    /**
     * Number of times the job may be attempted.
     */
    public int $tries;

    /**
     * Number of seconds to wait before retrying.
     */
    public int $backoff;

    /**
     * Maximum seconds the job can run.
     */
    public int $timeout;

    protected string $inputPath;
    protected string $outputPath;
    protected array $options;

    /**
     * Create a new job instance.
     */
    public function __construct(
        string $inputPath,
        string $outputPath,
        array $options = []
    ) {
        $this->inputPath = $inputPath;
        $this->outputPath = $outputPath;
        $this->options = $options;

        // Set from config
        $this->tries = config('gopdf.queue.tries', 3);
        $this->backoff = config('gopdf.queue.backoff', 30);
        $this->timeout = config('gopdf.timeout.single', 120);

        // Use configured queue
        $this->onQueue(config('gopdf.queue.queue', 'pdf-conversions'));
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        Log::info('GoPdf Job: Starting conversion', [
            'input' => $this->inputPath,
            'output' => $this->outputPath,
        ]);

        try {
            $result = PdfConverter::convert(
                $this->inputPath,
                $this->outputPath,
                $this->options
            );

            Log::info('GoPdf Job: Conversion completed', [
                'input' => $this->inputPath,
                'output' => $this->outputPath,
                'time_ms' => $result['process_time_ms'] ?? null,
            ]);

        } catch (PdfConversionException $e) {
            Log::error('GoPdf Job: Conversion failed', [
                'input' => $this->inputPath,
                'error' => $e->getMessage(),
                'code' => $e->getErrorCode(),
            ]);

            throw $e;
        }
    }

    /**
     * Handle job failure.
     */
    public function failed(\Throwable $exception): void
    {
        Log::error('GoPdf Job: Job failed permanently', [
            'input' => $this->inputPath,
            'output' => $this->outputPath,
            'error' => $exception->getMessage(),
        ]);
    }

    /**
     * Get the tags for the job (for Horizon).
     */
    public function tags(): array
    {
        return [
            'gopdf',
            'conversion',
            'file:' . basename($this->inputPath),
        ];
    }

    /**
     * Determine if the job should be marked as failed on timeout.
     */
    public function shouldMarkAsFailedOnTimeout(): bool
    {
        return true;
    }
}
