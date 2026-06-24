<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Wallet extends Model
{
    protected $fillable = ['balance', 'currency'];

    protected $casts = [
        'balance' => 'decimal:2',
    ];
}
