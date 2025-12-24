<?php

namespace NikunjKothiya\GoPdfConverter\Exceptions;

use Exception;

class PdfConversionException extends Exception
{
    protected string $errorCode;
    protected ?string $inputFile;
    protected ?string $details;

    public function __construct(
        string $message,
        string $errorCode = 'CONVERSION_FAILED',
        ?string $inputFile = null,
        ?string $details = null,
        int $code = 0,
        ?\Throwable $previous = null
    ) {
        parent::__construct($message, $code, $previous);
        $this->errorCode = $errorCode;
        $this->inputFile = $inputFile;
        $this->details = $details;
    }

    public function getErrorCode(): string
    {
        return $this->errorCode;
    }

    public function getInputFile(): ?string
    {
        return $this->inputFile;
    }

    public function getDetails(): ?string
    {
        return $this->details;
    }

    /**
     * Create exception from Go binary JSON output
     */
    public static function fromJson(array $data): self
    {
        $error = $data['error'] ?? [];
        
        return new self(
            $error['message'] ?? 'Conversion failed',
            $error['code'] ?? 'CONVERSION_FAILED',
            $error['file'] ?? null,
            $error['details'] ?? null
        );
    }

    /**
     * Get exception as array for API responses
     */
    public function toArray(): array
    {
        return [
            'error' => true,
            'code' => $this->errorCode,
            'message' => $this->getMessage(),
            'file' => $this->inputFile,
            'details' => $this->details,
        ];
    }
}

class FileNotFoundException extends PdfConversionException
{
    public function __construct(string $file)
    {
        parent::__construct(
            "File not found: {$file}",
            'FILE_NOT_FOUND',
            $file
        );
    }
}

class UnsupportedFormatException extends PdfConversionException
{
    public function __construct(string $file, string $format)
    {
        parent::__construct(
            "Unsupported file format: {$format}",
            'UNSUPPORTED_FORMAT',
            $file,
            "Supported formats: csv, xlsx, xls, pptx, ppt"
        );
    }
}

class BinaryNotFoundException extends PdfConversionException
{
    public function __construct(?string $path = null)
    {
        parent::__construct(
            "Go PDF converter binary not found" . ($path ? ": {$path}" : ""),
            'BINARY_NOT_FOUND',
            null,
            "Run 'php artisan gopdf:install' to install the binary"
        );
    }
}

class TimeoutException extends PdfConversionException
{
    public function __construct(string $file, int $timeout)
    {
        parent::__construct(
            "Conversion timed out after {$timeout} seconds",
            'TIMEOUT',
            $file
        );
    }
}

class CorruptFileException extends PdfConversionException
{
    public function __construct(string $file, ?string $details = null)
    {
        parent::__construct(
            "Input file appears to be corrupt or invalid",
            'CORRUPT_FILE',
            $file,
            $details
        );
    }
}
