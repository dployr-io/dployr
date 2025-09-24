<?php

use App\Http\Controllers\Projects\Services\BlueprintsController;

Route::middleware(['auth', 'verified'])->group(function() {
    Route::get('deployments', [BlueprintsController::class, 'index'])->name('deploymentsList');
 
    Route::get('deployments/{blueprint}', [BlueprintsController::class, 'show'])->name('deploymentsShow');
});