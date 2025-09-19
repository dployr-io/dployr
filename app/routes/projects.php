<?php

use App\Http\Controllers\Projects\ProjectsController;

Route::middleware(['auth', 'verified'])->group(function() {
    Route::get('projects', [ProjectsController::class, 'index'])->name('projectsList');

    Route::get('projects/{project}', [ProjectsController::class, 'show'])->name('projectsShow');

    Route::post('projects', [ProjectsController::class, 'store'])->name('projectsCreate');
});