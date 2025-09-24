<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Concerns\HasUlids;

class Service extends Model
{
    use HasUlids;

    /**
     * The attributes that are mass assignable.
     *
     * @var list<string>
     */
    protected $fillable = [
        'id',
        'name',
        'source',
        'runtime',
        'run_cmd',
        'port',
        'working_dir',
        'output_dir',
        'image',
        'spec',
        'env_vars',
        'secrets',
        'remote_id',
        'ci_remote_id',
    ];

    public function remote() 
    {
        return $this->belongsTo(Remote::class);
    }
}
