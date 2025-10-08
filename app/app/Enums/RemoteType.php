<?php

namespace App\Enums;

enum RemoteType {
    case GitHub;
    case GitLab;
    case BitBucket;
}