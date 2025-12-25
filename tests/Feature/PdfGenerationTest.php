<?php

namespace Tests\Feature;

use NikunjKothiya\GoPdfConverter\Tests\TestCase;
use NikunjKothiya\GoPdfConverter\Services\GoPdfService;
use Illuminate\Support\Facades\File;

class PdfGenerationTest extends TestCase
{
    protected $binaryPath;
    protected $inputFile;
    protected $outputDir;

    protected function setUp(): void
    {
        parent::setUp();
        
        // Ensure we reference the binary correctly
        $this->binaryPath = __DIR__ . '/../../go-binary/gopdfconv';
        
        if (!file_exists($this->binaryPath)) {
            $this->markTestSkipped("Binary not found at {$this->binaryPath}. Please run 'go build' in go-binary directory.");
        }

        // Use the sample fixture
        $this->inputFile = __DIR__ . '/../fixtures/sample_100.csv';
        $this->outputDir = __DIR__ . '/../outputs';

        if (!File::exists($this->outputDir)) {
            File::makeDirectory($this->outputDir, 0755, true);
        }
    }

    protected function getService(): GoPdfService
    {
        return new GoPdfService($this->binaryPath);
    }

    /**
     * Test 1: Basic Conversion
     * Verifies that the converter works with zero configuration, applying default styles.
     */
    public function test_basic_conversion_defaults()
    {
        $outputFile = $this->outputDir . '/01_basic_default.pdf';
        if (file_exists($outputFile)) unlink($outputFile);

        $this->getService()->csv($this->inputFile)
            ->toPdf($outputFile)
            ->convert();

        $this->assertFileExists($outputFile);
        $this->assertGreaterThan(1000, filesize($outputFile));
        
        echo "\n[Test 1] Generated Basic PDF: $outputFile";
    }

    /**
     * Test 2: Custom Header & Footer
     * Verifies custom text for header and footer, including page numbering alignment.
     */
    public function test_custom_header_footer()
    {
        $outputFile = $this->outputDir . '/02_custom_header_footer.pdf';
        if (file_exists($outputFile)) unlink($outputFile);

        $this->getService()->csv($this->inputFile)
            ->toPdf($outputFile)
            ->headerText("Monthly Sales Report - Confidential") // Center Aligned
            ->footerText("Internal Use Only")                   // Left Aligned (Page X of Y on Right)
            ->convert();

        $this->assertFileExists($outputFile);
        
        echo "\n[Test 2] Generated Custom Header/Footer PDF: $outputFile";
    }

    /**
     * Test 3: Styling & Branding
     * Verifies changing colors, grid lines, and fonts to match brand guidelines.
     */
    public function test_styling_and_branding()
    {
        $outputFile = $this->outputDir . '/03_styled_branding.pdf';
        if (file_exists($outputFile)) unlink($outputFile);

        $this->getService()->csv($this->inputFile)
            ->toPdf($outputFile)
            ->headerText("Styled Brand Report")
            ->headerColor("#4A90E2") // Custom Blue Header
            ->rowColor("#F5F7FA")    // Subtle Gray Alternating Rows
            ->borderColor("#D3D3D3") // Light Gray Borders
            ->showGridLines(true)
            ->convert();

        $this->assertFileExists($outputFile);
        
        echo "\n[Test 3] Generated Styled PDF: $outputFile";
    }

    /**
     * Test 4: Smart Layout Engine
     * Verifies Auto-Orientation and Weighted Column Compression.
     * The input CSV has wide text; this should trigger Landscape mode and compress the Description column.
     */
    public function test_smart_layout_engine()
    {
        $outputFile = $this->outputDir . '/04_smart_layout.pdf';
        if (file_exists($outputFile)) unlink($outputFile);

        $this->getService()->csv($this->inputFile)
            ->toPdf($outputFile)
            ->headerText("Smart Layout Analysis")
            ->autoOrientation() // Should detect width and switch to Landscape
            ->convert();

        $this->assertFileExists($outputFile);
        
        echo "\n[Test 4] Generated Smart Layout PDF: $outputFile";
    }

    /**
     * Test 5: Advanced Options
     * Verifies margins, watermarks, and specific page sizes.
     */
    public function test_advanced_watermark_options()
    {
        $outputFile = $this->outputDir . '/05_watermark.pdf';
        if (file_exists($outputFile)) unlink($outputFile);

        $this->getService()->csv($this->inputFile)
            ->toPdf($outputFile)
            ->headerText("Draft Version")
            ->watermarkText("DRAFT - CONFIDENTIAL", 0.15) // 15% Opacity
            ->margin(15)
            ->fontSize(9)
            ->convert();

        $this->assertFileExists($outputFile);
        
        echo "\n[Test 5] Generated Watermarked PDF: $outputFile";
    }
}
