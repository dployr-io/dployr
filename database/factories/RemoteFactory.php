<?php

namespace Database\Factories;

use Illuminate\Database\Eloquent\Factories\Factory;
use Symfony\Component\Uid\Ulid;

class RemoteFactory extends Factory
{
    /**
     * Define the model's default state.
     *
     * @return array<string, mixed>
     */
    public function definition(): array
    {
        return [
            'id' => Ulid::generate(),
            'name' => fake()->name(),
            'provider' => 'github.com',
            'repository' => fake()->name(),
            'branch' => fake()->name(),
        ];
    }
}
