<?php

namespace NikunjKothiya\GoPdfConverter\Tests\Unit;

use NikunjKothiya\GoPdfConverter\Tests\TestCase;
use NikunjKothiya\GoPdfConverter\Services\GoPdfService;
use NikunjKothiya\GoPdfConverter\PdfBuilder;
use NikunjKothiya\GoPdfConverter\BatchBuilder;

class GoPdfServiceTest extends TestCase
{
    protected GoPdfService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = new GoPdfService();
    }

    /** @test */
    public function it_returns_supported_formats()
    {
        $formats = $this->service->getSupportedFormats();

        $this->assertIsArray($formats);
        $this->assertContains('csv', $formats);
        $this->assertContains('xlsx', $formats);
        $this->assertContains('pptx', $formats);
    }

    /** @test */
    public function it_creates_pdf_builder_for_csv()
    {
        $builder = $this->service->csv('/path/to/file.csv');

        $this->assertInstanceOf(PdfBuilder::class, $builder);
        $this->assertEquals('/path/to/file.csv', $builder->getInputPath());
    }

    /** @test */
    public function it_creates_pdf_builder_for_excel()
    {
        $builder = $this->service->excel('/path/to/file.xlsx');

        $this->assertInstanceOf(PdfBuilder::class, $builder);
        $this->assertEquals('/path/to/file.xlsx', $builder->getInputPath());
    }

    /** @test */
    public function it_creates_pdf_builder_for_pptx()
    {
        $builder = $this->service->pptx('/path/to/file.pptx');

        $this->assertInstanceOf(PdfBuilder::class, $builder);
        $this->assertEquals('/path/to/file.pptx', $builder->getInputPath());
    }

    /** @test */
    public function it_creates_batch_builder()
    {
        $files = ['/path/to/file1.csv', '/path/to/file2.xlsx'];
        $builder = $this->service->batch($files);

        $this->assertInstanceOf(BatchBuilder::class, $builder);
        $this->assertEquals($files, $builder->getInputPaths());
        $this->assertEquals(2, $builder->count());
    }

    /** @test */
    public function it_has_xlsx_alias()
    {
        $builder1 = $this->service->xlsx('/path/to/file.xlsx');
        $builder2 = $this->service->excel('/path/to/file.xlsx');

        $this->assertInstanceOf(PdfBuilder::class, $builder1);
        $this->assertInstanceOf(PdfBuilder::class, $builder2);
    }

    /** @test */
    public function it_has_powerpoint_alias()
    {
        $builder1 = $this->service->pptx('/path/to/file.pptx');
        $builder2 = $this->service->powerpoint('/path/to/file.pptx');

        $this->assertInstanceOf(PdfBuilder::class, $builder1);
        $this->assertInstanceOf(PdfBuilder::class, $builder2);
    }
}
