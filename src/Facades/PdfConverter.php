<?php

namespace NikunjKothiya\GoPdfConverter\Facades;

use Illuminate\Support\Facades\Facade;
use NikunjKothiya\GoPdfConverter\Services\GoPdfService;

/**
 * @method static \NikunjKothiya\GoPdfConverter\PdfBuilder csv(string $inputPath)
 * @method static \NikunjKothiya\GoPdfConverter\PdfBuilder excel(string $inputPath)
 * @method static \NikunjKothiya\GoPdfConverter\PdfBuilder xlsx(string $inputPath)
 * @method static \NikunjKothiya\GoPdfConverter\PdfBuilder pptx(string $inputPath)
 * @method static \NikunjKothiya\GoPdfConverter\PdfBuilder powerpoint(string $inputPath)
 * @method static \NikunjKothiya\GoPdfConverter\PdfBuilder from(string $inputPath)
 * @method static \NikunjKothiya\GoPdfConverter\BatchBuilder batch(array $inputPaths)
 * @method static array convert(string $inputPath, string $outputPath, array $options = [])
 * @method static bool isAvailable()
 * @method static string getBinaryPath()
 * @method static array getSupportedFormats()
 * 
 * @see \NikunjKothiya\GoPdfConverter\Services\GoPdfService
 */
class PdfConverter extends Facade
{
    /**
     * Get the registered name of the component.
     */
    protected static function getFacadeAccessor(): string
    {
        return 'gopdf.converter';
    }
}
