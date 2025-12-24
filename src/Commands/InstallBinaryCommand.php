<?php

namespace NikunjKothiya\GoPdfConverter\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\File;
use Illuminate\Support\Facades\Process;

class InstallBinaryCommand extends Command
{
    /**
     * The name and signature of the console command.
     */
    protected $signature = 'gopdf:install
                            {--force : Force reinstall even if binary exists}
                            {--path= : Custom installation path}
                            {--with-libreoffice : Also install LibreOffice for PPT support}
                            {--skip-libreoffice : Skip LibreOffice installation prompt}';

    /**
     * The console command description.
     */
    protected $description = 'Install or update the Go PDF converter binary and dependencies';

    /**
     * GitHub release URL pattern
     */
    protected string $releaseUrl = 'https://github.com/nikunjkothiya/laravel-go-pdf-converter/releases/latest/download';

    /**
     * Execute the console command.
     */
    public function handle(): int
    {
        $this->info('╔═══════════════════════════════════════════════════════════════╗');
        $this->info('║        Go PDF Converter - Installation Wizard                 ║');
        $this->info('╚═══════════════════════════════════════════════════════════════╝');
        $this->newLine();

        // Step 1: Install the Go binary
        $this->info('Step 1/3: Installing Go PDF Binary...');
        $binaryInstalled = $this->installBinary();
        
        if (!$binaryInstalled) {
            return Command::FAILURE;
        }

        // Step 2: Check/Install LibreOffice for PPT support
        $this->newLine();
        $this->info('Step 2/3: Checking LibreOffice for PPT support...');
        $this->checkLibreOffice();

        // Step 3: Verify installation
        $this->newLine();
        $this->info('Step 3/3: Verifying installation...');
        $this->verifyInstallation();

        $this->newLine();
        $this->info('╔═══════════════════════════════════════════════════════════════╗');
        $this->info('║           Installation Complete! 🎉                           ║');
        $this->info('╚═══════════════════════════════════════════════════════════════╝');
        
        $this->showUsageExamples();

        return Command::SUCCESS;
    }

    /**
     * Install the Go binary
     */
    protected function installBinary(): bool
    {
        // Detect platform
        $os = $this->detectOS();
        $arch = $this->detectArch();

        $this->info("  Detected platform: {$os}-{$arch}");

        // Build binary name
        $binaryName = "gopdfconv-{$os}-{$arch}";
        if ($os === 'windows') {
            $binaryName .= '.exe';
        }

        // Determine install path
        $installPath = $this->option('path') ?? $this->getDefaultInstallPath();
        $binaryPath = $installPath . '/' . $binaryName;

        // Also create a generic symlink
        $genericPath = $installPath . '/gopdfconv';
        if ($os === 'windows') {
            $genericPath .= '.exe';
        }

        // Check if already installed
        if (file_exists($binaryPath) && !$this->option('force')) {
            $this->info("  ✓ Binary already installed at: {$binaryPath}");
            return true;
        }

        // Ensure directory exists
        if (!is_dir($installPath)) {
            mkdir($installPath, 0755, true);
        }

        // Check if binary is bundled with package
        $bundledPath = dirname(__DIR__, 2) . '/bin/' . $binaryName;
        $genericBundled = dirname(__DIR__, 2) . '/bin/gopdfconv';
        
        if (file_exists($bundledPath)) {
            $this->info("  Using bundled binary");
            if (copy($bundledPath, $binaryPath)) {
                chmod($binaryPath, 0755);
                $this->info("  ✓ Binary installed to: {$binaryPath}");
                return true;
            }
        }

        // Check if generic binary exists
        if (file_exists($genericBundled)) {
            $this->info("  Using bundled binary");
            if (copy($genericBundled, $binaryPath)) {
                chmod($binaryPath, 0755);
                $this->info("  ✓ Binary installed to: {$binaryPath}");
                return true;
            }
        }

        // Try to download from GitHub releases
        $this->info("  Downloading from GitHub...");
        
        $downloadUrl = "{$this->releaseUrl}/{$binaryName}";

        try {
            $response = Http::timeout(300)
                ->withOptions(['sink' => $binaryPath])
                ->get($downloadUrl);

            if ($response->successful()) {
                chmod($binaryPath, 0755);
                $this->info("  ✓ Binary downloaded and installed");
                return true;
            } else {
                $this->warn("  Download failed. HTTP Status: " . $response->status());
                $this->showManualBinaryInstructions($os, $arch);
                return false;
            }

        } catch (\Exception $e) {
            $this->warn("  Download failed: " . $e->getMessage());
            $this->showManualBinaryInstructions($os, $arch);
            return false;
        }
    }

    /**
     * Check and optionally install LibreOffice
     */
    protected function checkLibreOffice(): void
    {
        $libreOfficePath = $this->findLibreOffice();
        
        if ($libreOfficePath) {
            $this->info("  ✓ LibreOffice found: {$libreOfficePath}");
            $this->info("  → Full PPT/PPTX support is available!");
            return;
        }

        $this->warn("  ⚠ LibreOffice not found");
        $this->line("  → PPT files will have limited support (text-only extraction)");
        $this->line("  → PPTX files will work but with reduced fidelity");
        
        if ($this->option('skip-libreoffice')) {
            return;
        }

        // Ask user if they want to install
        if ($this->option('with-libreoffice') || $this->confirm('Do you want to install LibreOffice for full PPT support?')) {
            $this->installLibreOffice();
        } else {
            $this->showLibreOfficeInstructions();
        }
    }

    /**
     * Find LibreOffice installation
     */
    protected function findLibreOffice(): ?string
    {
        $paths = [
            '/usr/bin/libreoffice',
            '/usr/bin/soffice',
            '/usr/local/bin/libreoffice',
            '/Applications/LibreOffice.app/Contents/MacOS/soffice',
            'C:\\Program Files\\LibreOffice\\program\\soffice.exe',
            'C:\\Program Files (x86)\\LibreOffice\\program\\soffice.exe',
        ];

        foreach ($paths as $path) {
            if (file_exists($path)) {
                return $path;
            }
        }

        // Try which/where command
        $result = Process::run('which libreoffice 2>/dev/null || which soffice 2>/dev/null');
        if ($result->successful() && trim($result->output())) {
            return trim($result->output());
        }

        return null;
    }

    /**
     * Install LibreOffice
     */
    protected function installLibreOffice(): void
    {
        $os = $this->detectOS();

        $this->info("  Installing LibreOffice...");

        switch ($os) {
            case 'linux':
                $this->installLibreOfficeLinux();
                break;
            case 'darwin':
                $this->installLibreOfficeMac();
                break;
            case 'windows':
                $this->showWindowsLibreOfficeInstructions();
                break;
        }
    }

    /**
     * Install LibreOffice on Linux
     */
    protected function installLibreOfficeLinux(): void
    {
        // Detect package manager
        $packageManager = null;
        $installCommand = null;

        if (file_exists('/usr/bin/apt-get')) {
            $packageManager = 'apt';
            $installCommand = 'sudo apt-get update && sudo apt-get install -y libreoffice';
        } elseif (file_exists('/usr/bin/yum')) {
            $packageManager = 'yum';
            $installCommand = 'sudo yum install -y libreoffice';
        } elseif (file_exists('/usr/bin/dnf')) {
            $packageManager = 'dnf';
            $installCommand = 'sudo dnf install -y libreoffice';
        } elseif (file_exists('/usr/bin/pacman')) {
            $packageManager = 'pacman';
            $installCommand = 'sudo pacman -S --noconfirm libreoffice-fresh';
        }

        if (!$packageManager) {
            $this->warn("  Could not detect package manager");
            $this->showLibreOfficeInstructions();
            return;
        }

        $this->line("  Using {$packageManager} to install LibreOffice...");
        $this->line("  Command: {$installCommand}");
        $this->newLine();

        if ($this->confirm('  Run this command now? (requires sudo)', true)) {
            $this->info("  Running installation... (this may take a few minutes)");
            
            $result = Process::timeout(600)->run($installCommand);
            
            if ($result->successful()) {
                $this->info("  ✓ LibreOffice installed successfully!");
            } else {
                $this->error("  Installation failed: " . $result->errorOutput());
                $this->showLibreOfficeInstructions();
            }
        } else {
            $this->line("  Skipped. You can run manually:");
            $this->line("  {$installCommand}");
        }
    }

    /**
     * Install LibreOffice on Mac
     */
    protected function installLibreOfficeMac(): void
    {
        // Check for Homebrew
        $result = Process::run('which brew');
        
        if (!$result->successful()) {
            $this->warn("  Homebrew not found. Please install manually:");
            $this->line("  1. Visit: https://www.libreoffice.org/download/download/");
            $this->line("  2. Download and install the macOS version");
            return;
        }

        $this->line("  Using Homebrew to install LibreOffice...");
        $this->line("  Command: brew install --cask libreoffice");

        if ($this->confirm('  Run this command now?', true)) {
            $this->info("  Running installation... (this may take a few minutes)");
            
            $result = Process::timeout(600)->run('brew install --cask libreoffice');
            
            if ($result->successful()) {
                $this->info("  ✓ LibreOffice installed successfully!");
            } else {
                $this->error("  Installation failed: " . $result->errorOutput());
            }
        }
    }

    /**
     * Show Windows LibreOffice instructions
     */
    protected function showWindowsLibreOfficeInstructions(): void
    {
        $this->warn("  Automatic installation not available on Windows");
        $this->newLine();
        $this->line("  Please install LibreOffice manually:");
        $this->line("  1. Visit: https://www.libreoffice.org/download/download/");
        $this->line("  2. Download the Windows installer");
        $this->line("  3. Run the installer with default options");
        $this->newLine();
        $this->line("  After installation, add to your .env:");
        $this->line("  GOPDF_LIBREOFFICE_PATH=C:\\Program Files\\LibreOffice\\program\\soffice.exe");
    }

    /**
     * Show LibreOffice installation instructions
     */
    protected function showLibreOfficeInstructions(): void
    {
        $this->newLine();
        $this->info("  To install LibreOffice manually:");
        $this->newLine();
        
        $os = $this->detectOS();
        
        switch ($os) {
            case 'linux':
                $this->line("  Ubuntu/Debian: sudo apt-get install libreoffice");
                $this->line("  CentOS/RHEL:   sudo yum install libreoffice");
                $this->line("  Fedora:        sudo dnf install libreoffice");
                $this->line("  Arch:          sudo pacman -S libreoffice-fresh");
                break;
            case 'darwin':
                $this->line("  Homebrew: brew install --cask libreoffice");
                $this->line("  Or download from: https://www.libreoffice.org/download/");
                break;
            case 'windows':
                $this->line("  Download from: https://www.libreoffice.org/download/");
                break;
        }
        
        $this->newLine();
        $this->line("  After installation, set in .env (if not auto-detected):");
        $this->line("  GOPDF_LIBREOFFICE_PATH=/path/to/libreoffice");
    }

    /**
     * Verify the installation
     */
    protected function verifyInstallation(): void
    {
        $binaryPath = $this->findBinary();
        
        if (!$binaryPath) {
            $this->error("  ✗ Binary not found!");
            return;
        }

        // Test the binary
        $result = Process::run("{$binaryPath} --version");
        
        if ($result->successful()) {
            $version = trim($result->output());
            $this->info("  ✓ Binary working: {$version}");
        } else {
            $this->warn("  ⚠ Binary found but failed to run");
        }

        // Check supported formats
        $this->newLine();
        $this->info("  Supported formats:");
        
        $formats = [
            'CSV/TSV' => '✓ Native (fast)',
            'XLSX/XLS' => '✓ Native (fast)',
            'PPTX' => $this->findLibreOffice() ? '✓ Full fidelity (LibreOffice)' : '◐ Text extraction (native)',
            'PPT' => $this->findLibreOffice() ? '✓ Full fidelity (LibreOffice)' : '◐ Text extraction (native)',
        ];

        foreach ($formats as $format => $status) {
            $this->line("    {$format}: {$status}");
        }
    }

    /**
     * Find the installed binary
     */
    protected function findBinary(): ?string
    {
        $installPath = $this->getDefaultInstallPath();
        $os = $this->detectOS();
        $arch = $this->detectArch();
        
        $binaryName = "gopdfconv-{$os}-{$arch}";
        if ($os === 'windows') {
            $binaryName .= '.exe';
        }

        $paths = [
            $installPath . '/' . $binaryName,
            $installPath . '/gopdfconv',
            dirname(__DIR__, 2) . '/bin/' . $binaryName,
            dirname(__DIR__, 2) . '/bin/gopdfconv',
        ];

        foreach ($paths as $path) {
            if (file_exists($path) && is_executable($path)) {
                return $path;
            }
        }

        return null;
    }

    /**
     * Show usage examples
     */
    protected function showUsageExamples(): void
    {
        $this->newLine();
        $this->info('Usage Examples:');
        $this->newLine();
        
        $this->line('  // Using Facade');
        $this->line('  use NikunjKothiya\GoPdfConverter\Facades\PdfConverter;');
        $this->newLine();
        
        $this->line('  // Convert CSV to PDF');
        $this->line("  PdfConverter::csv('data.csv')->toPdf('output.pdf')->convert();");
        $this->newLine();
        
        $this->line('  // Convert Excel to PDF');
        $this->line("  PdfConverter::excel('report.xlsx')->toPdf('report.pdf')->convert();");
        $this->newLine();
        
        $this->line('  // Convert PowerPoint to PDF');
        $this->line("  PdfConverter::pptx('slides.pptx')->toPdf('slides.pdf')->convert();");
        $this->newLine();
        
        $this->line('  // Using Artisan command');
        $this->line('  php artisan pdf:convert input.csv output.pdf');
        $this->newLine();
    }

    /**
     * Show manual binary installation instructions
     */
    protected function showManualBinaryInstructions(string $os, string $arch): void
    {
        $this->newLine();
        $this->warn('Manual Binary Installation Required:');
        $this->newLine();
        $this->line('Option 1: Build from source');
        $this->line('  cd vendor/nikunjkothiya/laravel-go-pdf-converter/go-binary');
        $this->line('  go build -o ../bin/gopdfconv ./cmd/gopdfconv');
        $this->newLine();
        $this->line('Option 2: Download from GitHub releases');
        $this->line("  1. Visit: https://github.com/nikunjkothiya/laravel-go-pdf-converter/releases");
        $this->line("  2. Download: gopdfconv-{$os}-{$arch}" . ($os === 'windows' ? '.exe' : ''));
        $this->line('  3. Place in: vendor/nikunjkothiya/laravel-go-pdf-converter/bin/');
        $this->line('  4. Make executable: chmod +x <path>/gopdfconv-*');
        $this->newLine();
    }

    /**
     * Detect operating system
     */
    protected function detectOS(): string
    {
        $os = PHP_OS_FAMILY;

        if ($os === 'Windows') {
            return 'windows';
        } elseif ($os === 'Darwin') {
            return 'darwin';
        } else {
            return 'linux';
        }
    }

    /**
     * Detect CPU architecture
     */
    protected function detectArch(): string
    {
        $arch = php_uname('m');

        if (in_array($arch, ['x86_64', 'amd64', 'AMD64'])) {
            return 'amd64';
        } elseif (in_array($arch, ['aarch64', 'arm64', 'ARM64'])) {
            return 'arm64';
        } else {
            return 'amd64';
        }
    }

    /**
     * Get default installation path
     */
    protected function getDefaultInstallPath(): string
    {
        return dirname(__DIR__, 2) . '/bin';
    }
}
