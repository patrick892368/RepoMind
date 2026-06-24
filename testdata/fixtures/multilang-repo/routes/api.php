<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\OrderController;

Route::post('/order/create', [OrderController::class, 'create']);
Route::get('/wallet/info', [WalletController::class, 'info']);
