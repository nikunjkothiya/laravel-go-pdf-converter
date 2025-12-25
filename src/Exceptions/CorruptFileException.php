<?php

namespace NikunjKothiya\GoPdfConverter\Exceptions;

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
