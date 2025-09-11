<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Inertia\Inertia;

class DashboardController extends Controller
{
    public function index(Request $request)
    {
        return Inertia::render('Dashboard', [
            'projects' => Project::all()->map(function ($user) {
                return [
                    'id' => '20639a54-a504-45c2-b425-6d672303285e',
                    'name' => 'E-commerce Platform',
                    'remoteRepo' => 'https://github.com/company/ecommerce-platform.git',
                    'lastCommitMessage' => 'feat: add payment processing integration'
                ];
            }),
        ]);
    }
}
