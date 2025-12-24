<?php

return [
    /*
    |--------------------------------------------------------------------------
    | Go Binary Path
    |--------------------------------------------------------------------------
    |
    | The path to the gopdfconv binary. If null, the package will auto-detect
    | the appropriate binary from the bin/ directory based on your OS/arch.
    |
    */
    'binary_path' => env('GOPDF_BINARY_PATH', null),

    /*
    |--------------------------------------------------------------------------
    | LibreOffice Path
    |--------------------------------------------------------------------------
    |
    | Optional path to LibreOffice installation for high-fidelity PPTX
    | conversion. If null, the Go binary will auto-detect common locations.
    |
    */
    'libreoffice_path' => env('GOPDF_LIBREOFFICE_PATH', null),

    /*
    |--------------------------------------------------------------------------
    | Temporary Directory
    |--------------------------------------------------------------------------
    |
    | Directory for temporary files during conversion. Uses Laravel's
    | storage/app/temp by default.
    |
    */
    'temp_dir' => env('GOPDF_TEMP_DIR', storage_path('app/temp/gopdf')),

    /*
    |--------------------------------------------------------------------------
    | Default Page Settings
    |--------------------------------------------------------------------------
    |
    | Default page configuration for PDF output.
    |
    */
    'defaults' => [
        'page_size' => 'A4',        // A4, Letter, Legal, A3
        'orientation' => 'portrait', // portrait, landscape
        'margin' => 20,              // points
        'font_size' => 10,           // points
        'header_row' => true,        // Treat first row as header (CSV/Excel)
    ],

    /*
    |--------------------------------------------------------------------------
    | Timeout Settings
    |--------------------------------------------------------------------------
    |
    | Maximum time allowed for conversions (in seconds).
    |
    */
    'timeout' => [
        'single' => 120,    // Single file conversion
        'batch' => 600,     // Batch conversion
    ],

    /*
    |--------------------------------------------------------------------------
    | Queue Settings
    |--------------------------------------------------------------------------
    |
    | Configuration for queued conversions.
    |
    */
    'queue' => [
        'connection' => env('GOPDF_QUEUE_CONNECTION', null), // null = default
        'queue' => env('GOPDF_QUEUE_NAME', 'pdf-conversions'),
        'tries' => 3,
        'backoff' => 30,
    ],

    /*
    |--------------------------------------------------------------------------
    | Batch Processing
    |--------------------------------------------------------------------------
    |
    | Settings for batch file processing.
    |
    */
    'batch' => [
        'workers' => env('GOPDF_BATCH_WORKERS', 0), // 0 = auto (CPU cores)
        'chunk_size' => 10,                          // Files per batch chunk
    ],

    /*
    |--------------------------------------------------------------------------
    | Cleanup Settings
    |--------------------------------------------------------------------------
    |
    | Automatic cleanup of temporary files.
    |
    */
    'cleanup' => [
        'enabled' => true,
        'after_conversion' => true,  // Delete temp files after conversion
        'max_age_hours' => 24,       // Delete temp files older than this
    ],

    /*
    |--------------------------------------------------------------------------
    | Logging
    |--------------------------------------------------------------------------
    |
    | Conversion logging settings.
    |
    */
    'logging' => [
        'enabled' => env('GOPDF_LOGGING', true),
        'channel' => env('GOPDF_LOG_CHANNEL', null), // null = default
    ],
];
