<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Concerns\HasUlids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class Blueprint extends Model
{
    /** @use HasFactory<\Database\Factories\BlueprintFactory> */
    use HasFactory, HasUlids;

    /**
     * The attributes that are mass assignable.
     *
     * @var list<string>
     */
    protected $fillable = [
        'id',
        'config',
        'status', // e.g., 'pending', 'completed', 'in_progress', 'failed'
        'metadata',
    ];

    protected $casts = [
        'config' => 'array',
        'metadata' => 'array',
    ];

    public function services()
    {
        return $this->hasMany(Service::class);
    }

    public function getRemoteObjAttribute()
    {
        $remoteId = $this->config['remote'] ?? null;

        return $remoteId ? Remote::find($remoteId) : null;
    }

    public function getCiRemoteObjAttribute()
    {
        $ciRemoteId = $this->config['ci_remote'] ?? null;

        return $ciRemoteId ? Remote::find($ciRemoteId) : null;
    }
}
