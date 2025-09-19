<?php

use App\Http\Controllers\Projects\Remotes\RemotesController;

Route::middleware(['auth', 'verified'])->group(function() {
    Route::get('remotes', [RemotesController::class, 'index'])->name('remotesList');

    Route::get('remotes/{project}', [RemotesController::class, 'show'])->name('remotesShow');

    Route::post('remotes', [RemotesController::class, 'store'])->name('remotesCreate');
    
    Route::post('remotes/search', [RemotesController::class, 'search'])->name('remotesSearch');
});