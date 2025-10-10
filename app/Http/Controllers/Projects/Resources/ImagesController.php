<?php

namespace App\Http\Controllers\Projects\Resources;

use App\Http\Controllers\Controller;
use Inertia\Inertia;

class ImagesController extends Controller
{
    /**
     * Show all images page.
     */
    public function index()
    {
        return Inertia::render('resources/images');
    }
}
