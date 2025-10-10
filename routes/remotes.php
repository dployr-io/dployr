<?php

use App\Http\Controllers\Projects\Resources\RemotesController;

Route::middleware(['auth', 'verified'])->prefix('resources/remotes')->group(function () {
    // All remotes page
    Route::get('/', [RemotesController::class, 'index'])->name('remotesIndex');

    // Fetch details about a new remote url
    Route::post('/search', [RemotesController::class, 'search'])->name('remotesSearch');

    // JSON resources
    Route::get('/fetch', [RemotesController::class, 'fetch'])->name('remotesFetch');

    // View a single remote's page
    Route::get('/{remote}', [RemotesController::class, 'show'])->name('remotesShow');

    // Create remote
    Route::post('/', [RemotesController::class, 'store'])->name('remotesCreate');
});
