<?php

use App\Http\Controllers\System\SystemController;
use App\Http\Controllers\System\UpdateConfigController;

Route::middleware(['auth', 'verified'])->group(function () {
    Route::get('system', [SystemController::class, 'index'])->name('systemInfo');

    Route::post('system/config', [UpdateConfigController::class, 'store'])->name('updateConfig');
});
