<?php

namespace App\Http\Controllers\System;

use App\Models\Config;
use Illuminate\Http\Request;
use Illuminate\Http\RedirectResponse;
use Inertia\Controller;

class UpdateConfigController extends Controller
{
    /**
     * Update a config value. Creates new entry if key is missing.
     * @param \Illuminate\Http\Request $request
     * @return RedirectResponse
     */
    public function store(Request $request): RedirectResponse 
    {
        $request->validate([
            '*' => [function ($attribute, $value, $fail) use ($request) {
                if (
                    empty($request->github_token) &&
                    empty($request->gitlab_token) &&
                    empty($request->bitbucket_token) &&
                    empty($request->slack_api_key)
                ) {
                    $fail('At least one token or API key must be provided.');
                }
            }],
        ]);

        $keys = [
            'github_token',
            'gitlab_token',
            'bitbucket_token',
            'slack_api_key',
        ];

        foreach ($keys as $key) {
            if ($request->has($key)) {
                Config::updateOrCreate(
                    ['key' => $key],
                    ['value' => $request->input($key)]
                );
            }
        }

        return back()->with('success', __('Configuration updated successfully!'));
    }
}
