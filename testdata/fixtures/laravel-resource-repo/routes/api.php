<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\OrderController;
use App\Http\Controllers\WalletController;

Route::resource('/orders', OrderController::class);

Route::prefix('api/v1')->group(function () {
    Route::apiResource('wallets', WalletController::class);
});
