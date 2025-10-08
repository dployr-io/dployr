<?php

use Inertia\Inertia;

Route::middleware('auth')->group(function () {
    Route::get('console', function () {
        return Inertia::render('console/index');
    })->name('console');
});
