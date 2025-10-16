<?php

use App\Http\Controllers\Projects\Services\RuntimesController;
use Illuminate\Support\Facades\Route;

Route::middleware('auth')->group(function () {
    Route::get('runtimes', [RuntimesController::class, 'list'])->name('listRuntimes');
});
