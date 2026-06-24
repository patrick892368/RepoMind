<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\OrderController;
use App\Http\Controllers\WalletController;
use App\Http\Controllers\ReportController;

Route::resource('/orders', OrderController::class)->only(['index', 'show']);

Route::prefix('api/v1')->group(function () {
    Route::apiResource('wallets', WalletController::class)->except(['destroy']);
    Route::resource('reports', ReportController::class)->except('edit');
});
