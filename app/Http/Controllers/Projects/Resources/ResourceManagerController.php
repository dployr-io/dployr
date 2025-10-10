<?php

namespace App\Http\Controllers\Projects\Resources;

use App\Http\Controllers\Controller;
use Inertia\Inertia;

class ResourceManagerController extends Controller
{
    /**
     * Show resource manager page.
     */
    public function index()
    {
        return Inertia::render('resources/resource-manager');
    }
}
