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


