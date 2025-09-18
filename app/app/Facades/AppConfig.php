<?php

namespace App\Facades;

use Illuminate\Support\Facades\Facade;

class AppConfig extends Facade
{
    protected static function getFacadeAccessor()
    {
        return 'app.config';
    }
}