<?php

use App\Http\Controllers\Projects\ProjectsController;

Route::middleware(['auth', 'verified'])->group(function() {
    Route::get('projects', [ProjectsController::class, 'index'])->name('projectsList');
    
    Route::post('projects', [ProjectsController::class, 'store'])->name('projectsCreate');
    
    Route::post('projects/search', [ProjectsController::class, 'search'])->name('projectsSearch');
});