<?php

use App\Http\Controllers\Projects\Resources\SpecsController;

Route::middleware(['auth', 'verified'])->prefix('resources/specs')->group(function () {
    // All images page
    Route::get('/', [SpecsController::class, 'index'])->name('specsIndex');
});
