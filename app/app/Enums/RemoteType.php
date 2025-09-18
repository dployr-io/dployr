<?php

namespace App\Enums;

enum RemoteType {
    case Github;
    case Gitlab;
    case BitBucket;
}