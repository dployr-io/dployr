<?php

use App\Http\Controllers\Projects\Services\BlueprintsController;

Route::middleware(['auth', 'verified'])->prefix('deployments')->group(function () {
    // Show all blueprints page
    Route::get('/', [BlueprintsController::class, 'index'])->name('deploymentsIndex');

    // JSON blueprints
    Route::get('/fetch', [BlueprintsController::class, 'fetch'])->name('deploymentsList');

    // View a single blueprint's page
    Route::get('/{blueprint}', [BlueprintsController::class, 'show'])->name('deploymentsShow');
});
