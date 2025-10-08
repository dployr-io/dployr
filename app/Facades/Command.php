<?php

namespace App\Facades;

use Illuminate\Support\Facades\Facade;
use App\Services\CmdService;

class Command extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return CmdService::class;
    }
}
