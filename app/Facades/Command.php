<?php

namespace App\Facades;

use App\Services\Cmd;
use Illuminate\Support\Facades\Facade;

class Command extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return Cmd::class;
    }
}
