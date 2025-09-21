<?php

use App\Http\Controllers\Projects\Remotes\RemotesController;

Route::middleware(['auth', 'verified'])->group(function() {
    Route::get('resources/remotes', [RemotesController::class, 'index'])->name('remotesList');
 
    // It is very important this route comes before remoteShow to avoid URL collision
    Route::post('resources/remotes/search', [RemotesController::class, 'search'])->name('remotesSearch');

    Route::get('resources/remotes/{project}', [RemotesController::class, 'show'])->name('remotesShow');

    Route::post('resources/remotes', [RemotesController::class, 'store'])->name('remotesCreate');
});