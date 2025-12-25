<?php

namespace NikunjKothiya\GoPdfConverter\Exceptions;

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
