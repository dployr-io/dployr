<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Config extends Model
{
    protected $fillable = ['key', 'value', 'description'];
    
    protected $casts = [
        'value' => 'encrypted'
    ];

    protected static function boot()
    {
        parent::boot();
        
        // Auto-refresh config singleton on model changes
        static::saved(function () {
            app()->forgetInstance('app.config');
        });
        
        static::deleted(function () {
            app()->forgetInstance('app.config');
        });
    }
}
