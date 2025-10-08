<?php

namespace App\Enums;

enum JobStatus: string
{
    case PENDING = 'pending';

    case IN_PROGRESS = 'in_progress';

    case FAILED = 'failed';

    case COMPLETED = 'completed';
}
