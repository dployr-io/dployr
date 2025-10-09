<?php

use App\Http\Controllers\Logs\LogsController;
use Illuminate\Support\Facades\Route;
use Inertia\Inertia;

Route::middleware('auth')->group(function () {
    Route::get('logs', fn () => Inertia::render('logs/index'))->name('logs');

    Route::get('logs/stream', [LogsController::class, 'stream'])->name('logsStream');
});
