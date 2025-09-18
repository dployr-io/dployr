<?php

namespace App\Enums;

enum ConfigKey: string
{
    case GITHUB_TOKEN = 'github_token';
    case GITLAB_TOKEN = 'gitlab_token';
}
