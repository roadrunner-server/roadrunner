<?php
declare(strict_types=1);
/**
 * RoadRunner.
 *
 * @license   MIT
 * @author    Anton Titov (Wolfy-J)
 */

namespace Spiral\RoadRunner\QuickBuild;

final class Builder
{
    const DOCKER = 'spiralscout/rr-build';

    /**
     * Coloring.
     *
     * @var array
     */
    protected static $colors = [
        "reset"  => "\e[0m",
        "white"  => "\033[1;38m",
        "red"    => "\033[0;31m",
        "green"  => "\033[0;32m",
        "yellow" => "\033[1;93m",
        "gray"   => "\033[0;90m"
    ];

    /** @var array */
    private $config;

    /**
     * @param array $config
     */
    protected function __construct(array $config)
    {
        $this->config = $config;
    }

    /**
     * Validate the build configuration.
     *
     * @return array
     */
    public function configErrors(): array
    {
        $errors = [];
        if (!isset($this->config["commands"])) {
            $errors[] = "Directive 'commands' missing";
        }

        if (!isset($this->config["packages"])) {
            $errors[] = "Directive 'packages' missing";
        }

        if (!isset($this->config["register"])) {
            $errors[] = "Directive 'register' missing";
        }

        return $errors;
    }

    /**
     * Build the application.
     *
     * @param string $directory
     * @param string $template
     * @param string $output
     * @param string $version
     */
    public function build(string $directory, string $template, string $output, string $version)
    {
        $filename = $directory . "/main.go";
        $output = $output . ($this->getOS() == 'windows' ? '.exe' : '');

        // step 1, generate template
        $this->generate($template, $filename);

        $command = sprintf(
            'docker run --rm -v "%s":/mnt -e RR_VERSION=%s -e GOARCH=amd64 -e GOOS=%s %s /bin/bash -c "mv /mnt/main.go main.go; bash compile.sh; cp rr /mnt/%s;"',
            $directory,
            $version,
            $this->getOS(),
            self::DOCKER,
            $output
        );

        self::cprintf("<yellow>%s</reset>\n", $command);

        // run the build
        $this->run($command, true);

        if (!file_exists($directory . '/' . $output)) {
            self::cprintf("<red>Build has failed!</reset>");
            return;
        }

        self::cprintf("<green>Build complete!</reset>\n");
        $this->run($directory . '/' . $output, false);
    }

    /**
     * @param string $command
     * @param bool   $shadow
     */
    protected function run(string $command, bool $shadow = false)
    {
        $shadow && self::cprintf("<gray>");
        passthru($command);
        $shadow && self::cprintf("</reset>");
    }

    /**
     * @param string $template
     * @param string $filename
     */
    protected function generate(string $template, string $filename)
    {
        $body = file_get_contents($template);

        $replace = [
            '// -packages- //' => '"' . join("\"\n\"", $this->config['packages']) . '"',
            '// -commands- //' => '_ "' . join("\"\n_ \"", $this->config['commands']) . '"',
            '// -register- //' => join("\n", $this->config['register'])
        ];

        // compile the template
        $result = str_replace(array_keys($replace), array_values($replace), $body);
        file_put_contents($filename, $result);
    }

    /**
     * @return string
     */
    protected function getOS(): string
    {
        $os = strtolower(PHP_OS);

        if (strpos($os, 'darwin') !== false) {
            return 'darwin';
        }

        if (strpos($os, 'win') !== false) {
            return 'windows';
        }

        return "linux";
    }

    /**
     * Create new builder using given config.
     *
     * @param string $config
     * @return Builder|null
     */
    public static function loadConfig(string $config): ?Builder
    {
        if (!file_exists($config)) {
            return null;
        }

        $configData = json_decode(file_get_contents($config), true);
        if (!is_array($configData)) {
            return null;
        }

        return new Builder($configData);
    }

    /**
     * Make colored output.
     *
     * @param string $format
     * @param mixed  ...$args
     */
    public static function cprintf(string $format, ...$args)
    {
        if (self::isColorsSupported()) {
            $format = preg_replace_callback("/<\/?([^>]+)>/", function ($value) {
                return self::$colors[$value[1]];
            }, $format);
        } else {
            $format = preg_replace("/<[^>]+>/", "", $format);
        }

        echo sprintf($format, ...$args);
    }

    /**
     * @return bool
     */
    public static function isWindows(): bool
    {
        return \DIRECTORY_SEPARATOR === '\\';
    }

    /**
     * Returns true if the STDOUT supports colorization.
     *
     * @codeCoverageIgnore
     * @link https://github.com/symfony/Console/blob/master/Output/StreamOutput.php#L94
     * @param mixed $stream
     * @return bool
     */
    public static function isColorsSupported($stream = STDOUT): bool
    {
        if ('Hyper' === getenv('TERM_PROGRAM')) {
            return true;
        }

        try {
            if (\DIRECTORY_SEPARATOR === '\\') {
                return (
                        function_exists('sapi_windows_vt100_support')
                        && @sapi_windows_vt100_support($stream)
                    ) || getenv('ANSICON') !== false
                    || getenv('ConEmuANSI') == 'ON'
                    || getenv('TERM') == 'xterm';
            }

            if (\function_exists('stream_isatty')) {
                return (bool)@stream_isatty($stream);
            }

            if (\function_exists('posix_isatty')) {
                return (bool)@posix_isatty($stream);
            }

            $stat = @fstat($stream);
            // Check if formatted mode is S_IFCHR
            return $stat ? 0020000 === ($stat['mode'] & 0170000) : false;
        } catch (\Throwable $e) {
            return false;
        }
    }
}
