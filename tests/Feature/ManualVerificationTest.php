<?php

namespace NikunjKothiya\GoPdfConverter\Tests\Feature;

use NikunjKothiya\GoPdfConverter\Tests\TestCase;
use NikunjKothiya\GoPdfConverter\Services\GoPdfService;

class ManualVerificationTest extends TestCase
{
    /** @test */
    public function it_generates_pdf_for_manual_verification()
    {
        // 1. Setup paths
        $binaryPath = __DIR__ . '/../../go-binary/gopdfconv';
        if (!file_exists($binaryPath)) {
            $this->markTestSkipped("Binary not found at $binaryPath. Run 'go build' first.");
        }

        $inputFile = __DIR__ . '/../fixtures/sample_100.csv';
        $outputFile = __DIR__ . '/../../output_sample.pdf';

        if (file_exists($outputFile)) {
            unlink($outputFile);
        }

        // 2. Initialize Service with explicit binary path
        $service = new GoPdfService($binaryPath);

        // 3. Run Conversion
        $result = $service->csv($inputFile)
            ->toPdf($outputFile)
            ->headerText('Verification Report')
            ->footerText('Sensitive & Confidential')
            ->autoOrientation()
            ->convert();

        // 4. Verification
        $this->assertTrue($result['success']);
        $this->assertFileExists($outputFile);
        
        echo "\n[ManualVerification] PDF created at: $outputFile\n";
    }
}
