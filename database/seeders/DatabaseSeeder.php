<?php

namespace Database\Seeders;

use App\Enums\ConfigKey;
use App\Models\Config;
use App\Models\User;
// use Illuminate\Database\Console\Seeds\WithoutModelEvents;
use Illuminate\Database\Seeder;

class DatabaseSeeder extends Seeder
{
    /**
     * Seed the application's database.
     */
    public function run(): void
    {
        foreach (ConfigKey::cases() as $case) {
            Config::firstOrCreate(
                ['key' => $case->value],
                ['value' => null]
            );
        }
    }
}
