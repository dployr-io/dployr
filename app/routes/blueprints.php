<?php

use App\Http\Controllers\Projects\Services\BlueprintsController;

Route::middleware(['auth', 'verified'])->group(function() {
    Route::get('projects/services/deployments', [BlueprintsController::class, 'index'])->name('deploymentsList');
 
    Route::get('projects/services/deployments/{deployment}', [BlueprintsController::class, 'show'])->name('deploymentsShow');
});