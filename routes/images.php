<?php

use App\Http\Controllers\Projects\Resources\ImagesController;

Route::middleware(['auth', 'verified'])->prefix('resources/images')->group(function () {
    // All images page
    Route::get('/', [ImagesController::class, 'index'])->name('imagesIndex');
});
