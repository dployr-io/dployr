<?php

namespace App\Constants;

final readonly class Runtimes
{
    public const STATIC = 'static';

    public const GO = 'go';

    public const PHP = 'php';

    public const PYTHON = 'python';

    public const NODE_JS = 'node-js';

    public const RUBY = 'ruby';

    public const DOTNET = 'dotnet';

    public const JAVA = 'java';

    public const DOCKER = 'docker';

    public const K3S = 'k3s';

    public const STANDALONE_RUNTIMES = [
        self::GO,
        self::PHP,
        self::PYTHON,
        self::NODE_JS,
        self::RUBY,
        self::DOTNET,
        self::JAVA,
    ];

    public const CONTAINER_RUNTIMES = [
        self::DOCKER,
        self::K3S,
    ];

    public const RUNTIMES = [
        self::STATIC,
        ...self::STANDALONE_RUNTIMES,
        ...self::CONTAINER_RUNTIMES,
    ];
}
