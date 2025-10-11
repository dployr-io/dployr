<?php

/**
 * Release script that bumps version number and creates a new git tag.
 * Usage: composer release [-v|--version VERSION] [-m|--message MESSAGE]
 * If version is not provided, automatically bumps patch version number.
 * This script expects a VERSION file in the project's root.
 */

function showHelp() {
    echo <<<HELP
Usage: composer release [OPTIONS]

Options:
  -v, --version            Specify exact version (e.g., 2.5.0)
  -m, --message            Commit message (default: "Version bump")
  -h, --help               Show this help message

Examples:
  composer release                                   # Bump patch (1.2.3 -> 1.2.4)
  composer release -- -v 2.0.0                       # Set specific version
  composer release -- -m "feat: New feature"         # With custom message
  composer release -- -v 2.5.0 -m "Custom message"   # Both options

HELP;
    exit(0);
}

$options = [
    'version' => null,
    'message' => 'Version bump',
];

$args = array_slice($argv, 1);
for ($i = 0; $i < count($args); $i++) {
    $arg = $args[$i];
    
    if ($arg === '-h' || $arg === '--help') {
        showHelp();
    } elseif ($arg === '-v' || $arg === '--version') {
        if (!isset($args[$i + 1])) {
            echo "Error: --version requires a value\n\n";
            showHelp();
        }
        $options['version'] = $args[++$i];
    } elseif ($arg === '-m' || $arg === '--message') {
        if (!isset($args[$i + 1])) {
            echo "Error: --message requires a value\n\n";
            showHelp();
        }
        $options['message'] = $args[++$i];
    } else {
        echo "Error: Unknown option '$arg'\n\n";
        showHelp();
    }
}
if (!file_exists('VERSION')) {
    echo "Error: VERSION file not found in current directory\n";
    exit(1);
}

$currentVersion = trim(file_get_contents('VERSION'));

if ($options['version']) {
    $newVersion = $options['version'];
} else {
    list($major, $minor, $patch) = explode('.', $currentVersion);
    $newVersion = $major . '.' . $minor . '.' . ($patch + 1);
}
if (!preg_match('/^\d+\.\d+\.\d+$/', $newVersion)) {
    echo "Error: Version must be in format X.Y.Z\n";
    exit(1);
}

list($newMajor, $newMinor, $newPatch) = explode('.', $newVersion);

if ($newMajor > 9 || $newMinor > 99 || $newPatch > 999) {
    echo "Error: Version cannot exceed 9.99.999\n";
    exit(1);
}
if (version_compare($newVersion, $currentVersion, '<=')) {
    echo "Error: New version must be greater than current version $currentVersion\n";
    exit(1);
}

file_put_contents('VERSION', $newVersion . PHP_EOL);
echo "Bumped version from $currentVersion to $newVersion\n";

exec('git add -A');
exec('git commit -m ' . escapeshellarg($options['message']));
exec('git push');

echo "Successfully released version $newVersion\n";