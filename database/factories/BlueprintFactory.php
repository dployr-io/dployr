<?php

namespace Database\Factories;

use Illuminate\Database\Eloquent\Factories\Factory;
use Symfony\Component\Uid\Ulid;

class BlueprintFactory extends Factory
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
            'config' => json_encode([
                'id' => Ulid::generate(),
                'name' => $this->faker->name(),
                'source' => 'remote',
            ]),
            'status' => 'pending',
        ];
    }
}
