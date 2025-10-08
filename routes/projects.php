<?php

use App\Http\Controllers\Projects\ProjectsController;

Route::middleware(['auth', 'verified'])->prefix('projects')->group(function() {
    // Show all projects page
    Route::get('/', [ProjectsController::class, 'index'])->name('projectsIndex');

    // JSON projects
    Route::get('/fetch', [ProjectsController::class, 'fetch'])->name('projectsList'); 

    // View a single project's page
    Route::get('/{project}', [ProjectsController::class, 'show'])->name('projectsShow');

    // Create project
    Route::post('/', [ProjectsController::class, 'store'])->name('projectsCreate');
});