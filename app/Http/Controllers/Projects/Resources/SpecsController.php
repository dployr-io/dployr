<?php

namespace App\Http\Controllers\Projects\Resources;

use App\Http\Controllers\Controller;
use Inertia\Inertia;

class SpecsController extends Controller
{
    /**
     * Show all specs page.
     */
    public function index()
    {
        return Inertia::render('resources/specs');
    }
}
