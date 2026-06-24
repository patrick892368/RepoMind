<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\OrderController;
use App\Http\Controllers\WalletController;

Route::prefix('api/v1')->group(function () {
    Route::post('/orders', [OrderController::class, 'store']);
    Route::middleware('auth:sanctum')->prefix('wallet')->group(function () {
        Route::get('/info', [WalletController::class, 'info']);
    });
});

Route::group(['prefix' => 'admin'], function () {
    Route::delete('/orders/{id}', [OrderController::class, 'destroy']);
});
