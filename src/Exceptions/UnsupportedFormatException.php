<?php

namespace NikunjKothiya\GoPdfConverter\Exceptions;

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
