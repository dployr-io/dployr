<?php

use App\Http\Controllers\Projects\Resources\ResourceManagerController;

Route::middleware(['auth', 'verified'])->prefix('resources/manager')->group(function () {
    Route::get('/', [ResourceManagerController::class, 'index'])->name('resourceManagerIndex');
});
