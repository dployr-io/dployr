<?php

namespace App\Providers;

use App\Models\Config;
use Illuminate\Support\ServiceProvider;

class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
        $this->app->singleton('app.config', function () {
            try {
                return Config::all()->pluck('value', 'key');
            } catch (\Exception $e) {
                // Handle case when table doesn't exist (migrations not run)
                return collect();
            }
        });
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        $version = trim(file_get_contents(base_path('VERSION')));
        config(['app.version' => $version]);

        if (config('app.env') !== 'production') {
            \Http::globalOptions([
                'verify' => false,
            ]);
        }
    }
}
