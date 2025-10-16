<?php

use Illuminate\Support\Facades\Route;
use Inertia\Inertia;

Route::get('/', fn () => Inertia::render('welcome'))->name('home');

require __DIR__.'/projects.php';
require __DIR__.'/resources.php';
require __DIR__.'/services.php';
require __DIR__.'/blueprints.php';
require __DIR__.'/remotes.php';
require __DIR__.'/runtimes.php';
require __DIR__.'/images.php';
require __DIR__.'/specs.php';
require __DIR__.'/notifications.php';
require __DIR__.'/settings.php';
require __DIR__.'/auth.php';
require __DIR__.'/logs.php';
require __DIR__.'/console.php';
require __DIR__.'/system.php';
