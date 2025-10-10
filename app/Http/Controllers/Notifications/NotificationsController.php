<?php

namespace App\Http\Controllers\Notifications;

use App\Http\Controllers\Controller;
use Inertia\Inertia;

class NotificationsController extends Controller
{
    /**
     * Show all notifications page.
     */
    public function index()
    {
        return Inertia::render('notifications/index');
    }
}
