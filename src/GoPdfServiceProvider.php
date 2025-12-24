<?php

namespace NikunjKothiya\GoPdfConverter;

use Illuminate\Support\ServiceProvider;
use NikunjKothiya\GoPdfConverter\Services\GoPdfService;
use NikunjKothiya\GoPdfConverter\Commands\ConvertCommand;
use NikunjKothiya\GoPdfConverter\Commands\InstallBinaryCommand;

class GoPdfServiceProvider extends ServiceProvider
{
    /**
     * Register services.
     */
    public function register(): void
    {
        // Merge config
        $this->mergeConfigFrom(
            __DIR__ . '/../config/gopdf.php',
            'gopdf'
        );

        // Register the main service
        $this->app->singleton('gopdf.converter', function ($app) {
            return new GoPdfService(
                config('gopdf.binary_path'),
                config('gopdf.libreoffice_path'),
                config('gopdf.temp_dir'),
                config('gopdf.defaults'),
                config('gopdf.timeout')
            );
        });

        // Register alias
        $this->app->alias('gopdf.converter', GoPdfService::class);
    }

    /**
     * Bootstrap services.
     */
    public function boot(): void
    {
        // Publish config
        if ($this->app->runningInConsole()) {
            $this->publishes([
                __DIR__ . '/../config/gopdf.php' => config_path('gopdf.php'),
            ], 'gopdf-config');

            // Register commands
            $this->commands([
                ConvertCommand::class,
                InstallBinaryCommand::class,
            ]);
        }

        // Ensure temp directory exists
        $tempDir = config('gopdf.temp_dir');
        if ($tempDir && !is_dir($tempDir)) {
            @mkdir($tempDir, 0755, true);
        }
    }

    /**
     * Get the services provided by the provider.
     */
    public function provides(): array
    {
        return ['gopdf.converter', GoPdfService::class];
    }
}
