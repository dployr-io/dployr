<?php

use Illuminate\Support\Facades\Route;
use Inertia\Inertia;
use App\Http\Controllers\Projects\Services\LogsController;

Route::middleware('auth')->group(function () {
    Route::get('logs', fn() => Inertia::render('logs/index'))->name('logs');
    
    Route::get('logs/stream', [LogsController::class, 'logs'])->name('logsStream');
});
