<?php

namespace NikunjKothiya\GoPdfConverter\Tests\Unit;

use NikunjKothiya\GoPdfConverter\Tests\TestCase;
use NikunjKothiya\GoPdfConverter\Exceptions\PdfConversionException;
use NikunjKothiya\GoPdfConverter\Exceptions\FileNotFoundException;
use NikunjKothiya\GoPdfConverter\Exceptions\UnsupportedFormatException;
use NikunjKothiya\GoPdfConverter\Exceptions\BinaryNotFoundException;
use NikunjKothiya\GoPdfConverter\Exceptions\TimeoutException;
use NikunjKothiya\GoPdfConverter\Exceptions\CorruptFileException;

class ExceptionTest extends TestCase
{
    /** @test */
    public function pdf_conversion_exception_has_error_code()
    {
        $exception = new PdfConversionException(
            'Test message',
            'TEST_ERROR',
            '/path/to/file',
            'Additional details'
        );

        $this->assertEquals('TEST_ERROR', $exception->getErrorCode());
        $this->assertEquals('/path/to/file', $exception->getInputFile());
        $this->assertEquals('Additional details', $exception->getDetails());
    }

    /** @test */
    public function pdf_conversion_exception_converts_to_array()
    {
        $exception = new PdfConversionException(
            'Test message',
            'TEST_ERROR',
            '/path/to/file',
            'Details here'
        );

        $array = $exception->toArray();

        $this->assertTrue($array['error']);
        $this->assertEquals('TEST_ERROR', $array['code']);
        $this->assertEquals('Test message', $array['message']);
        $this->assertEquals('/path/to/file', $array['file']);
        $this->assertEquals('Details here', $array['details']);
    }

    /** @test */
    public function pdf_conversion_exception_creates_from_json()
    {
        $data = [
            'error' => [
                'code' => 'JSON_ERROR',
                'message' => 'Error from JSON',
                'file' => '/json/file.csv',
                'details' => 'JSON details',
            ],
        ];

        $exception = PdfConversionException::fromJson($data);

        $this->assertEquals('JSON_ERROR', $exception->getErrorCode());
        $this->assertEquals('Error from JSON', $exception->getMessage());
        $this->assertEquals('/json/file.csv', $exception->getInputFile());
        $this->assertEquals('JSON details', $exception->getDetails());
    }

    /** @test */
    public function file_not_found_exception_sets_proper_values()
    {
        $exception = new FileNotFoundException('/path/to/missing.csv');

        $this->assertEquals('FILE_NOT_FOUND', $exception->getErrorCode());
        $this->assertEquals('/path/to/missing.csv', $exception->getInputFile());
        $this->assertStringContains('File not found', $exception->getMessage());
    }

    /** @test */
    public function unsupported_format_exception_includes_format()
    {
        $exception = new UnsupportedFormatException('/path/to/file.doc', 'doc');

        $this->assertEquals('UNSUPPORTED_FORMAT', $exception->getErrorCode());
        $this->assertStringContains('doc', $exception->getMessage());
        $this->assertStringContains('Supported formats', $exception->getDetails());
    }

    /** @test */
    public function binary_not_found_exception_provides_help()
    {
        $exception = new BinaryNotFoundException('/custom/path');

        $this->assertEquals('BINARY_NOT_FOUND', $exception->getErrorCode());
        $this->assertStringContains('gopdf:install', $exception->getDetails());
    }

    /** @test */
    public function timeout_exception_includes_duration()
    {
        $exception = new TimeoutException('/path/to/file.csv', 120);

        $this->assertEquals('TIMEOUT', $exception->getErrorCode());
        $this->assertStringContains('120', $exception->getMessage());
    }

    /** @test */
    public function corrupt_file_exception_has_proper_code()
    {
        $exception = new CorruptFileException('/path/to/corrupt.xlsx', 'Invalid ZIP structure');

        $this->assertEquals('CORRUPT_FILE', $exception->getErrorCode());
        $this->assertEquals('/path/to/corrupt.xlsx', $exception->getInputFile());
        $this->assertEquals('Invalid ZIP structure', $exception->getDetails());
    }

    /**
     * Helper to check if string contains substring
     */
    protected function assertStringContains(string $needle, string $haystack): void
    {
        $this->assertTrue(
            str_contains($haystack, $needle),
            "Failed asserting that '$haystack' contains '$needle'"
        );
    }
}
