<?php

use App\Http\Controllers\Notifications\NotificationsController;

Route::middleware(['auth', 'verified'])->prefix('notifications')->group(function () {
    Route::get('/', [NotificationsController::class, 'index'])->name('notificationsIndex');
});
