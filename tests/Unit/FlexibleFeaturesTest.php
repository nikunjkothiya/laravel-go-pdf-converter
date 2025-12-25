<?php

namespace NikunjKothiya\GoPdfConverter\Tests\Unit;

use NikunjKothiya\GoPdfConverter\Tests\TestCase;
use NikunjKothiya\GoPdfConverter\PdfBuilder;
use NikunjKothiya\GoPdfConverter\Services\GoPdfService;

class FlexibleFeaturesTest extends TestCase
{
    protected GoPdfService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = new GoPdfService();
    }

    /** @test */
    public function it_sets_simplified_header_and_footer()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->headerText('Centered Header')
                ->footerText('Left Footer');

        $options = $builder->getOptions();
        $this->assertEquals('Centered Header', $options['header_text']);
        $this->assertEquals('Left Footer', $options['footer_text']);
    }

    /** @test */
    public function it_sets_auto_orientation()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        
        $builder->autoOrientation();
        $this->assertTrue($builder->getOptions()['auto_orientation']);

        $builder->autoOrientation(false);
        $this->assertFalse($builder->getOptions()['auto_orientation']);
    }

    /** @test */
    public function it_handles_all_options_together()
    {
        $builder = new PdfBuilder($this->service, '/input.csv');
        $builder->headerText('Confidential')
                ->footerText('Copyright 2025')
                ->autoOrientation(true);

        $options = $builder->getOptions();
        $this->assertEquals('Confidential', $options['header_text']);
        $this->assertEquals('Copyright 2025', $options['footer_text']);
        $this->assertTrue($options['auto_orientation']);
    }
}
