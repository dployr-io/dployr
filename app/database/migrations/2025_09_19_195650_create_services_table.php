<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create('services', function (Blueprint $table) {
            $table->ulid('id')->primary();
            $table->string('name');
            $table->string('source');
            $table->string('runtime');
            $table->string('run_cmd');
            $table->integer('port');
            $table->string('working_dir')->nullable();
            $table->string('output_dir')->nullable();
            $table->string('image')->nullable();
            $table->json('spec')->nullable();
            $table->json('env_vars')->nullable();
            $table->json('secrets')->nullable();
            $table->foreignUlid('project_id')->nullable()->constrained('projects')->nullOnDelete();
            $table->foreignUlid('remote_id')->nullable()->constrained('remotes')->nullOnDelete();
            $table->foreignUlid('ci_remote_id')->nullable()->constrained('remotes')->nullOnDelete();
            $table->timestamps();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('services');
    }
};
