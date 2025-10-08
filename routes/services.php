<?php

use App\Http\Controllers\Projects\Services\ServicesController;

Route::middleware(['auth', 'verified'])->group(function() {
    // Project's services page
    Route::get('projects/{project}/services', [ServicesController::class, 'deploy'])->name('servicesIndex');

    // Fetch all services for a given project
    Route::get('projects/{project}/services/fetch', [ServicesController::class, 'fetch'])->name('servicesList');

    // View a single project's page
    Route::get('projects/{project}/services/{service}', [ServicesController::class, 'show'])->name('servicesShow');

    // Create a service
    Route::post('projects/{project}/services', [ServicesController::class, 'store'])->name('servicesCreate');

    // Check if a port is available
    Route::post('projects/services/check-port', [ServicesController::class, 'checkPort'])->name('servicesCheckPort');
});