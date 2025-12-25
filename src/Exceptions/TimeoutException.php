<?php

namespace NikunjKothiya\GoPdfConverter\Exceptions;

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
