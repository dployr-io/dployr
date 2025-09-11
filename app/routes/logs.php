<?php

use Illuminate\Support\Facades\Route;
use Inertia\Inertia;

Route::middleware('auth')->group(function () {
    Route::get('logs', function () {
        return Inertia::render('logs/index');
    })->name('logs');
});
