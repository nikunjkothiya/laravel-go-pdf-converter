<?php

namespace NikunjKothiya\GoPdfConverter\Tests;

use Orchestra\Testbench\TestCase as BaseTestCase;
use NikunjKothiya\GoPdfConverter\GoPdfServiceProvider;

abstract class TestCase extends BaseTestCase
{
    protected function getPackageProviders($app): array
    {
        return [
            GoPdfServiceProvider::class,
        ];
    }

    protected function getPackageAliases($app): array
    {
        return [
            'PdfConverter' => \NikunjKothiya\GoPdfConverter\Facades\PdfConverter::class,
        ];
    }

    protected function getEnvironmentSetUp($app): void
    {
        // Setup default config
        $app['config']->set('gopdf.binary_path', null);
        $app['config']->set('gopdf.temp_dir', sys_get_temp_dir() . '/gopdf-test');
        $app['config']->set('gopdf.defaults', [
            'page_size' => 'A4',
            'orientation' => 'portrait',
            'margin' => 20,
            'font_size' => 10,
            'header_row' => true,
        ]);
    }

    /**
     * Get path to test fixtures
     */
    protected function getFixturePath(string $filename): string
    {
        return __DIR__ . '/fixtures/' . $filename;
    }

    /**
     * Get temp output path
     */
    protected function getTempOutputPath(string $filename): string
    {
        $dir = sys_get_temp_dir() . '/gopdf-test-output';
        if (!is_dir($dir)) {
            mkdir($dir, 0755, true);
        }
        return $dir . '/' . $filename;
    }

    /**
     * Clean up temp files after test
     */
    protected function tearDown(): void
    {
        parent::tearDown();

        // Clean up temp output directory
        $dir = sys_get_temp_dir() . '/gopdf-test-output';
        if (is_dir($dir)) {
            $files = glob($dir . '/*');
            foreach ($files as $file) {
                unlink($file);
            }
        }
    }
}
