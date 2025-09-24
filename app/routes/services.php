<?php

use App\Http\Controllers\Projects\Services\ServicesController;

Route::middleware(['auth', 'verified'])->group(function() {
    Route::get('projects/{project}/services', [ServicesController::class, 'deploy'])->name('servicesList');

    Route::get('projects/{project}/services/{service}', [ServicesController::class, 'show'])->name('servicesShow');

    Route::post('projects/{project}/services', [ServicesController::class, 'store'])->name('servicesCreate');
});