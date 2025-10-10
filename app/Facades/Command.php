<?php

namespace App\Facades;

use App\Services\CmdService;
use Illuminate\Support\Facades\Facade;

class Command extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return CmdService::class;
    }
}
