<?php

namespace NikunjKothiya\GoPdfConverter\Tests\Unit;

use NikunjKothiya\GoPdfConverter\Tests\TestCase;
use NikunjKothiya\GoPdfConverter\PdfBuilder;
use NikunjKothiya\GoPdfConverter\Services\GoPdfService;

class PdfBuilderTest extends TestCase
{
    protected GoPdfService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = new GoPdfService();
    }

    /** @test */
    public function it_sets_output_path()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->toPdf('/output.pdf');

        $this->assertEquals('/output.pdf', $builder->getOutputPath());
    }

    /** @test */
    public function it_has_save_to_alias()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->saveTo('/output.pdf');

        $this->assertEquals('/output.pdf', $builder->getOutputPath());
    }

    /** @test */
    public function it_sets_format()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->format('xlsx');

        $options = $builder->getOptions();
        $this->assertEquals('xlsx', $options['format']);
    }

    /** @test */
    public function it_sets_page_size()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->pageSize('Letter');

        $options = $builder->getOptions();
        $this->assertEquals('Letter', $options['page_size']);
    }

    /** @test */
    public function it_has_page_size_shortcuts()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        
        $builder->a4();
        $this->assertEquals('A4', $builder->getOptions()['page_size']);

        $builder->letter();
        $this->assertEquals('Letter', $builder->getOptions()['page_size']);

        $builder->legal();
        $this->assertEquals('Legal', $builder->getOptions()['page_size']);
    }

    /** @test */
    public function it_sets_orientation()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        
        $builder->landscape();
        $this->assertEquals('landscape', $builder->getOptions()['orientation']);

        $builder->portrait();
        $this->assertEquals('portrait', $builder->getOptions()['orientation']);
    }

    /** @test */
    public function it_sets_margin()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->margin(25.5);

        $options = $builder->getOptions();
        $this->assertEquals(25.5, $options['margin']);
    }

    /** @test */
    public function it_sets_font_size()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->fontSize(12);

        $options = $builder->getOptions();
        $this->assertEquals(12, $options['font_size']);
    }

    /** @test */
    public function it_sets_header_row_option()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        
        $builder->withHeaders();
        $this->assertTrue($builder->getOptions()['header_row']);

        $builder->withoutHeaders();
        $this->assertFalse($builder->getOptions()['header_row']);
    }

    /** @test */
    public function it_sets_timeout()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->timeout(300);

        $options = $builder->getOptions();
        $this->assertEquals(300, $options['timeout']);
    }

    /** @test */
    public function it_adds_custom_options()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->option('custom_key', 'custom_value');

        $options = $builder->getOptions();
        $this->assertEquals('custom_value', $options['custom_key']);
    }

    /** @test */
    public function it_merges_multiple_options()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->options([
            'page_size' => 'A3',
            'margin' => 30,
        ]);

        $options = $builder->getOptions();
        $this->assertEquals('A3', $options['page_size']);
        $this->assertEquals(30, $options['margin']);
    }

    /** @test */
    public function it_supports_method_chaining()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        
        $result = $builder
            ->toPdf('/output.pdf')
            ->pageSize('Letter')
            ->landscape()
            ->margin(25)
            ->fontSize(11)
            ->withHeaders()
            ->timeout(120);

        $this->assertInstanceOf(PdfBuilder::class, $result);
        $this->assertEquals('/output.pdf', $result->getOutputPath());
        
        $options = $result->getOptions();
        $this->assertEquals('Letter', $options['page_size']);
        $this->assertEquals('landscape', $options['orientation']);
        $this->assertEquals(25, $options['margin']);
        $this->assertEquals(11, $options['font_size']);
        $this->assertTrue($options['header_row']);
        $this->assertEquals(120, $options['timeout']);
    }

    /** @test */
    public function it_generates_output_path_from_input()
    {
        $builder = new PdfBuilder($this->service, '/path/to/document.xlsx');
        
        // getOutputPath returns null before conversion, but convert() would auto-generate
        $this->assertNull($builder->getOutputPath());
        $this->assertEquals('/path/to/document.xlsx', $builder->getInputPath());
    }
}
