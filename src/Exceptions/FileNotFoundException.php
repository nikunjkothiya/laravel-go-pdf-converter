<?php

namespace NikunjKothiya\GoPdfConverter\Exceptions;

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
