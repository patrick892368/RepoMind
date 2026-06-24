package apiroute

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patrick892368/RepoMind/internal/ir"
	"github.com/patrick892368/RepoMind/internal/scanner"
)

func TestExtractAPIRoutesFromFixture(t *testing.T) {
	root := filepath.Join("..", "..", "..", "testdata", "fixtures", "api-repo")
	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "ANY", "/login/", "views.login_view", "django")
	assertRoute(t, routes, "ANY", "/order/create/", "views.create_order", "django")
	assertRoute(t, routes, "POST", "/login", "login", "fastapi")
	assertRoute(t, routes, "GET", "/wallet/info", "wallet_info", "fastapi")
	assertRoute(t, routes, "POST", "/order/create", "createOrder", "express")
	assertRoute(t, routes, "GET", "/wallet/info", "walletController.info", "express")
	assertRoute(t, routes, "POST", "/order/create", "create", "nestjs")
	assertRoute(t, routes, "GET", "/order/status", "status", "nestjs")
}

func TestExtractPHPJavaAndGoRoutes(t *testing.T) {
	root := filepath.Join("..", "..", "..", "testdata", "fixtures", "multilang-repo")
	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "POST", "/order/create", "OrderController@create", "laravel")
	assertRoute(t, routes, "POST", "/order/create", "create", "spring")
	assertRoute(t, routes, "POST", "/login", "loginHandler", "go")
	assertRoute(t, routes, "GET", "/wallet/info", "walletInfo", "go")
}

func TestParseLaravelRouteGroupsAndControllerArrayHandlers(t *testing.T) {
	content := `<?php

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

Route::post('/login', [AuthController::class, 'login']);
`
	routes := parseLaravel("routes/api.php", content)

	assertRoute(t, routes, "POST", "/api/v1/orders", "OrderController@store", "laravel")
	assertRoute(t, routes, "GET", "/api/v1/wallet/info", "WalletController@info", "laravel")
	assertRoute(t, routes, "DELETE", "/admin/orders/{id}", "OrderController@destroy", "laravel")
	assertRoute(t, routes, "POST", "/login", "AuthController@login", "laravel")
	assertNoRoute(t, routes, "POST", "/orders", "OrderController@store", "laravel")
	assertNoRoute(t, routes, "GET", "/wallet/info", "WalletController@info", "laravel")
}

func TestParseLaravelResourceRoutes(t *testing.T) {
	content := `<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\OrderController;
use App\Http\Controllers\WalletController;

Route::resource('/orders', OrderController::class);

Route::prefix('api/v1')->group(function () {
    Route::apiResource('wallets', WalletController::class);
});
`
	routes := parseLaravel("routes/api.php", content)

	assertRoute(t, routes, "GET", "/orders", "OrderController@index", "laravel")
	assertRoute(t, routes, "POST", "/orders", "OrderController@store", "laravel")
	assertRoute(t, routes, "GET", "/orders/create", "OrderController@create", "laravel")
	assertRoute(t, routes, "GET", "/orders/{order}", "OrderController@show", "laravel")
	assertRoute(t, routes, "PUT", "/orders/{order}", "OrderController@update", "laravel")
	assertRoute(t, routes, "PATCH", "/orders/{order}", "OrderController@update", "laravel")
	assertRoute(t, routes, "DELETE", "/orders/{order}", "OrderController@destroy", "laravel")
	assertRoute(t, routes, "GET", "/orders/{order}/edit", "OrderController@edit", "laravel")
	assertRoute(t, routes, "GET", "/api/v1/wallets", "WalletController@index", "laravel")
	assertRoute(t, routes, "GET", "/api/v1/wallets/{wallet}", "WalletController@show", "laravel")
	assertRoute(t, routes, "PATCH", "/api/v1/wallets/{wallet}", "WalletController@update", "laravel")
	assertNoRoute(t, routes, "GET", "/api/v1/wallets/create", "WalletController@create", "laravel")
}

func TestParseLaravelResourceRouteOptions(t *testing.T) {
	content := `<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\OrderController;
use App\Http\Controllers\WalletController;
use App\Http\Controllers\ReportController;

Route::resource('/orders', OrderController::class)->only(['index', 'show']);

Route::prefix('api/v1')->group(function () {
    Route::apiResource('wallets', WalletController::class)->except(['destroy']);
    Route::resource('reports', ReportController::class)->except('edit');
});
`
	routes := parseLaravel("routes/api.php", content)

	assertRoute(t, routes, "GET", "/orders", "OrderController@index", "laravel")
	assertRoute(t, routes, "GET", "/orders/{order}", "OrderController@show", "laravel")
	assertNoRoute(t, routes, "POST", "/orders", "OrderController@store", "laravel")
	assertNoRoute(t, routes, "DELETE", "/orders/{order}", "OrderController@destroy", "laravel")
	assertRoute(t, routes, "PATCH", "/api/v1/wallets/{wallet}", "WalletController@update", "laravel")
	assertNoRoute(t, routes, "DELETE", "/api/v1/wallets/{wallet}", "WalletController@destroy", "laravel")
	assertRoute(t, routes, "GET", "/api/v1/reports/create", "ReportController@create", "laravel")
	assertNoRoute(t, routes, "GET", "/api/v1/reports/{report}/edit", "ReportController@edit", "laravel")
}

func TestParseLaravelResourceRouteParameters(t *testing.T) {
	content := `<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\OrderController;
use App\Http\Controllers\WalletController;

Route::resource('/orders', OrderController::class)->parameters(['orders' => 'order_uuid'])->only(['show', 'update']);

Route::prefix('api/v1')->group(function () {
    Route::apiResource('wallets', WalletController::class)->parameters(['wallets' => 'wallet_slug'])->except(['destroy']);
});
`
	routes := parseLaravel("routes/api.php", content)

	assertRoute(t, routes, "GET", "/orders/{order_uuid}", "OrderController@show", "laravel")
	assertRoute(t, routes, "PUT", "/orders/{order_uuid}", "OrderController@update", "laravel")
	assertRoute(t, routes, "PATCH", "/orders/{order_uuid}", "OrderController@update", "laravel")
	assertNoRoute(t, routes, "GET", "/orders/{order}", "OrderController@show", "laravel")
	assertNoRoute(t, routes, "POST", "/orders", "OrderController@store", "laravel")
	assertRoute(t, routes, "GET", "/api/v1/wallets/{wallet_slug}", "WalletController@show", "laravel")
	assertNoRoute(t, routes, "DELETE", "/api/v1/wallets/{wallet_slug}", "WalletController@destroy", "laravel")
}

func TestParseLaravelMultilineResourceRouteChains(t *testing.T) {
	content := `<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\OrderController;
use App\Http\Controllers\WalletController;

Route::resource('/orders', OrderController::class)
    ->parameters(['orders' => 'order_uuid'])
    ->only(['show', 'update']);

Route::prefix('api/v1')->group(function () {
    Route::apiResource('wallets', WalletController::class)
        ->parameters(['wallets' => 'wallet_slug'])
        ->except(['destroy']);
});
`
	routes := parseLaravel("routes/api.php", content)

	assertRoute(t, routes, "GET", "/orders/{order_uuid}", "OrderController@show", "laravel")
	assertRoute(t, routes, "PUT", "/orders/{order_uuid}", "OrderController@update", "laravel")
	assertRoute(t, routes, "PATCH", "/orders/{order_uuid}", "OrderController@update", "laravel")
	assertNoRoute(t, routes, "GET", "/orders/{order}", "OrderController@show", "laravel")
	assertNoRoute(t, routes, "POST", "/orders", "OrderController@store", "laravel")
	assertRoute(t, routes, "GET", "/api/v1/wallets/{wallet_slug}", "WalletController@show", "laravel")
	assertRoute(t, routes, "PATCH", "/api/v1/wallets/{wallet_slug}", "WalletController@update", "laravel")
	assertNoRoute(t, routes, "DELETE", "/api/v1/wallets/{wallet_slug}", "WalletController@destroy", "laravel")
}

func TestParseSymfonyAttributeRoutes(t *testing.T) {
	content := `<?php

namespace App\Controller;

use Symfony\Component\Routing\Attribute\Route;

#[Route('/blog')]
class BlogController
{
    #[Route('/', name: 'blog_index', methods: ['GET'])]
    #[Route('/rss.xml', name: 'blog_rss', methods: ['GET'])]
    public function index(): Response
    {
    }

    #[Route('/posts/{slug:post}', name: 'blog_post', methods: ['GET'])]
    public function postShow(): Response
    {
    }

    #[Route('/comment/{postSlug}/new', name: 'comment_new', methods: ['POST'])]
    public function commentNew(): Response
    {
    }
}
`
	routes := parseSymfony("src/Controller/BlogController.php", content)

	assertRoute(t, routes, "GET", "/blog", "index", "symfony")
	assertRoute(t, routes, "GET", "/blog/rss.xml", "index", "symfony")
	assertRoute(t, routes, "GET", "/blog/posts/{slug}", "postShow", "symfony")
	assertRoute(t, routes, "POST", "/blog/comment/{postSlug}/new", "commentNew", "symfony")
	assertNoRoute(t, routes, "GET", "/blog/posts/{slug:post}", "postShow", "symfony")
}

func TestParseNextJSAppRouterRoutes(t *testing.T) {
	content := `import { NextRequest } from "next/server";

export async function GET(request: NextRequest) {
  return Response.json({});
}

export const POST = async (request: NextRequest) => {
  return Response.json({});
};
`
	routes := parseNextJS("apps/web/app/api/stripe/checkout/route.ts", content)

	assertRoute(t, routes, "GET", "/api/stripe/checkout", "GET", "nextjs")
	assertRoute(t, routes, "POST", "/api/stripe/checkout", "POST", "nextjs")
}

func TestParseNextJSDynamicAppRouterRoute(t *testing.T) {
	content := `export async function DELETE() {
  return Response.json({});
}
`
	routes := parseNextJS("app/api/orders/[orderId]/route.ts", content)

	assertRoute(t, routes, "DELETE", "/api/orders/{orderId}", "DELETE", "nextjs")
}

func TestParseNextJSPagesAPIRoutes(t *testing.T) {
	content := `export default async function handler(req, res) {
  if (req.method === "POST") {
    return res.status(201).json({});
  }

  switch (req.method) {
  case "GET":
    return res.status(200).json({});
  }
}
`
	routes := parseNextJS("pages/api/orders/[orderId].ts", content)

	assertRoute(t, routes, "POST", "/api/orders/{orderId}", "handler", "nextjs")
	assertRoute(t, routes, "GET", "/api/orders/{orderId}", "handler", "nextjs")
}

func TestExtractNextJSRoutes(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "app/api/team/route.ts", `export async function GET() {
  return Response.json({});
}
`)
	writeRouteFile(t, root, "pages/api/orders/[orderId].ts", `export default async function handler(req, res) {
  if (req.method === "PATCH") {
    return res.status(200).json({});
  }
}
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/api/team", "GET", "nextjs")
	assertRoute(t, routes, "PATCH", "/api/orders/{orderId}", "handler", "nextjs")
}

func TestParseSpringMappingArraysAndRequestMethods(t *testing.T) {
	content := `package com.example;

import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping({"/api/v1", "/internal"})
public class OrderController {
    @GetMapping({"/orders", "/purchases"})
    public String list() {
        return "ok";
    }

    @RequestMapping(path = {"/orders/{id}", "/purchases/{id}"}, method = {RequestMethod.PUT, RequestMethod.PATCH})
    public String update() {
        return "ok";
    }

    @PostMapping(path = "/orders")
    public String create() {
        return "ok";
    }
}
`
	routes := parseSpring("src/main/java/com/example/OrderController.java", content)

	assertRoute(t, routes, "GET", "/api/v1/orders", "list", "spring")
	assertRoute(t, routes, "GET", "/api/v1/purchases", "list", "spring")
	assertRoute(t, routes, "GET", "/internal/orders", "list", "spring")
	assertRoute(t, routes, "PUT", "/api/v1/orders/{id}", "update", "spring")
	assertRoute(t, routes, "PATCH", "/api/v1/purchases/{id}", "update", "spring")
	assertRoute(t, routes, "POST", "/api/v1/orders", "create", "spring")
	assertRoute(t, routes, "POST", "/internal/orders", "create", "spring")
}

func TestParseGoRoutesWithASTHandlers(t *testing.T) {
	content := `package main

func register(router Router, controller OrderController, wallet WalletController) {
	router.POST("/order/create", controller.Create)
	router.GET("wallet/info", wallet.Info)
	router.Any("/health", healthHandler)
	router.GET("/inline", func(c Context) {})
	router.DELETE("/admin", requireAuth(adminHandler))
	router.PATCH("/order/:id", middleware.Require(controller.Update))
}
`
	routes := parseGoRoutes("cmd/api/main.go", content)

	assertRoute(t, routes, "POST", "/order/create", "controller.Create", "go")
	assertRoute(t, routes, "GET", "/wallet/info", "wallet.Info", "go")
	assertRoute(t, routes, "ANY", "/health", "healthHandler", "go")
	assertRoute(t, routes, "GET", "/inline", "inline", "go")
	assertRoute(t, routes, "DELETE", "/admin", "adminHandler", "go")
	assertRoute(t, routes, "PATCH", "/order/:id", "controller.Update", "go")
}

func TestParseGoRoutesWithNetHTTPMux(t *testing.T) {
	content := `package main

import "net/http"

func register(mux *http.ServeMux, router Router) {
	http.HandleFunc("/login", login)
	mux.HandleFunc("GET /wallet/info", walletInfo)
	mux.Handle("POST /order/create", http.HandlerFunc(createOrder))
	mux.Handle("/metrics", metricsHandler)
	mux.HandleFunc("DELETE /order/{id}", requireAuth(deleteOrder))
	mux.Handle("GET /inline", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	router.Handle("PUT", "/order/{id}", updateOrder)
}
`
	routes := parseGoRoutes("cmd/api/main.go", content)

	assertRoute(t, routes, "ANY", "/login", "login", "go")
	assertRoute(t, routes, "GET", "/wallet/info", "walletInfo", "go")
	assertRoute(t, routes, "POST", "/order/create", "createOrder", "go")
	assertRoute(t, routes, "ANY", "/metrics", "metricsHandler", "go")
	assertRoute(t, routes, "DELETE", "/order/{id}", "deleteOrder", "go")
	assertRoute(t, routes, "GET", "/inline", "inline", "go")
	assertRoute(t, routes, "PUT", "/order/{id}", "updateOrder", "go")
	assertNoRoute(t, routes, "ANY", "/GET", "/wallet/info", "go")
}

func TestParseGoRoutesWithChiRoutePrefix(t *testing.T) {
	content := `package main

func register(r Router, order OrderController) {
	r.Route("/api", func(r Router) {
		r.Get("/users", listUsers)
		r.Route("/orders", func(r Router) {
			r.Post("/", order.Create)
		})
	})
}
`
	routes := parseGoRoutes("cmd/api/main.go", content)

	assertRoute(t, routes, "GET", "/api/users", "listUsers", "go")
	assertRoute(t, routes, "POST", "/api/orders", "order.Create", "go")
}

func TestParseGoRoutesWithMountedSubrouterVariable(t *testing.T) {
	content := `package main

func register(r Router) {
	api := NewRouter()
	api.Get("/users", listUsers)
	api.Post("/orders", createOrder)
	r.Mount("/api", api)
}
`
	routes := parseGoRoutes("cmd/api/main.go", content)

	assertRoute(t, routes, "GET", "/api/users", "listUsers", "go")
	assertRoute(t, routes, "POST", "/api/orders", "createOrder", "go")
	assertNoRoute(t, routes, "GET", "/users", "listUsers", "go")
	assertNoRoute(t, routes, "POST", "/orders", "createOrder", "go")
}

func TestExtractGoSamePackageRouteFactoryPrefix(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "api/server.go", `package api

func register(r Router) {
	r.Mount("/api", orderRoutes())
}
`)
	writeRouteFile(t, root, "api/orders.go", `package api

func orderRoutes() Router {
	r := NewRouter()
	r.Get("/orders", listOrders)
	r.Post("/orders", createOrder)
	return r
}
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/api/orders", "listOrders", "go")
	assertRoute(t, routes, "POST", "/api/orders", "createOrder", "go")
	assertNoRoute(t, routes, "GET", "/orders", "listOrders", "go")
}

func TestExtractGoUnresolvedRouteFactoryKeepsChildRoute(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "api/server.go", `package api

func register(r Router) {
	r.Mount("/api", missingRoutes())
}
`)
	writeRouteFile(t, root, "api/orders.go", `package api

func orderRoutes() Router {
	r := NewRouter()
	r.Get("/orders", listOrders)
	return r
}
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/orders", "listOrders", "go")
	assertNoRoute(t, routes, "GET", "/api/orders", "listOrders", "go")
}

func TestExtractGoMethodRouteFactoryPrefix(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "api/server.go", `package api

func register(r Router) {
	r.Mount("/users", usersResource{}.Routes())
}
`)
	writeRouteFile(t, root, "api/users.go", `package api

type usersResource struct{}

func (rs usersResource) Routes() Router {
	r := NewRouter()
	r.Get("/", rs.List)
	r.Get("/{id}", rs.Get)
	return r
}
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/users", "rs.List", "go")
	assertRoute(t, routes, "GET", "/users/{id}", "rs.Get", "go")
	assertNoRoute(t, routes, "GET", "/{id}", "rs.Get", "go")
}

func TestExtractGoImportedPackageRouteFactoryPrefix(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "cmd/api/server.go", `package api

import "example.com/repomind/internal/users"

func register(r Router) {
	r.Mount("/api", users.Routes())
}
`)
	writeRouteFile(t, root, "internal/users/routes.go", `package users

func Routes() Router {
	r := NewRouter()
	r.Get("/users", listUsers)
	r.Post("/users", createUser)
	return r
}
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/api/users", "listUsers", "go")
	assertRoute(t, routes, "POST", "/api/users", "createUser", "go")
	assertNoRoute(t, routes, "GET", "/users", "listUsers", "go")
	assertNoRoute(t, routes, "POST", "/users", "createUser", "go")
}

func TestParseDjangoURLsWithSameFileIncludePrefix(t *testing.T) {
	content := `from django.urls import include, path
from . import views

api_patterns = [
    path("login/", views.login_view, name="login"),
    path("orders/create/", views.create_order, name="create_order"),
]

urlpatterns = [
    path("api/v1/", include(api_patterns)),
    path("health/", views.health, name="health"),
]
`
	routes := parseDjangoURLs("project/urls.py", content)

	assertRoute(t, routes, "ANY", "/api/v1/login/", "views.login_view", "django")
	assertRoute(t, routes, "ANY", "/api/v1/orders/create/", "views.create_order", "django")
	assertRoute(t, routes, "ANY", "/health/", "views.health", "django")
	assertNoRoute(t, routes, "ANY", "/login/", "views.login_view", "django")
}

func TestExtractDjangoModuleIncludePrefix(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "project/urls.py", `from django.urls import include, path

urlpatterns = [
    path("api/v1/", include("orders.urls")),
]
`)
	writeRouteFile(t, root, "orders/urls.py", `from django.urls import path
from . import views

urlpatterns = [
    path("create/", views.create_order, name="create_order"),
    path("status/", views.order_status, name="order_status"),
]
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "ANY", "/api/v1/create/", "views.create_order", "django")
	assertRoute(t, routes, "ANY", "/api/v1/status/", "views.order_status", "django")
	assertNoRoute(t, routes, "ANY", "/create/", "views.create_order", "django")
}

func TestParseDjangoRESTFrameworkRouterPrefix(t *testing.T) {
	content := `from django.urls import include, path
from rest_framework import routers
from . import views

router = routers.DefaultRouter()
router.register(r"users", views.UserViewSet, basename="user")

urlpatterns = [
    path("api/", include(router.urls)),
]
`
	routes := parseDjangoURLs("project/urls.py", content)

	assertRoute(t, routes, "GET", "/api/users/", "views.UserViewSet.list", "django")
	assertRoute(t, routes, "POST", "/api/users/", "views.UserViewSet.create", "django")
	assertRoute(t, routes, "GET", "/api/users/{id}/", "views.UserViewSet.retrieve", "django")
	assertRoute(t, routes, "DELETE", "/api/users/{id}/", "views.UserViewSet.destroy", "django")
}

func TestExtractDjangoRESTFrameworkCustomActions(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "project/urls.py", `from django.urls import include, path
from rest_framework import routers
from . import views

router = routers.DefaultRouter()
router.register(r"users", views.UserViewSet, basename="user")

urlpatterns = [
    path("api/", include(router.urls)),
]
`)
	writeRouteFile(t, root, "project/views.py", `from rest_framework.decorators import action
from rest_framework.viewsets import ModelViewSet

class UserViewSet(ModelViewSet):
    @action(detail=True, methods=["post"], url_path="set-password")
    def set_password(self, request, pk=None):
        pass

    @action(
        detail=False,
        methods=["get", "post"],
        url_path="recent",
    )
    def recent_users(self, request):
        pass
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "POST", "/api/users/{id}/set-password/", "views.UserViewSet.set_password", "django")
	assertRoute(t, routes, "GET", "/api/users/recent/", "views.UserViewSet.recent_users", "django")
	assertRoute(t, routes, "POST", "/api/users/recent/", "views.UserViewSet.recent_users", "django")
}

func TestExtractDjangoRESTFrameworkCustomActionsThroughModuleInclude(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "project/urls.py", `from django.urls import include, path

urlpatterns = [
    path("api/v1/", include("users.urls")),
]
`)
	writeRouteFile(t, root, "users/urls.py", `from django.urls import include, path
from rest_framework import routers
from . import views

router = routers.DefaultRouter()
router.register(r"users", views.UserViewSet, basename="user")

urlpatterns = [
    path("", include(router.urls)),
]
`)
	writeRouteFile(t, root, "users/views.py", `from rest_framework.decorators import action
from rest_framework.viewsets import ModelViewSet

class UserViewSet(ModelViewSet):
    @action(detail=True, methods=["post"], url_path="set-password")
    def set_password(self, request, pk=None):
        pass
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/api/v1/users/", "views.UserViewSet.list", "django")
	assertRoute(t, routes, "POST", "/api/v1/users/{id}/set-password/", "views.UserViewSet.set_password", "django")
	assertNoRoute(t, routes, "POST", "/users/{id}/set-password/", "views.UserViewSet.set_password", "django")
}

func TestParseFastAPIWithRouterPrefixes(t *testing.T) {
	content := `from fastapi import APIRouter, FastAPI

app = FastAPI()
router = APIRouter(prefix="/orders")

@router.post("/{order_id}/pay")
def pay_order(order_id: str):
    return {}

app.include_router(router, prefix="/api/v1")
`
	routes := parseFastAPI("app/api/routes.py", content)

	assertRoute(t, routes, "POST", "/api/v1/orders/{order_id}/pay", "pay_order", "fastapi")
}

func TestParseFastAPIWithMultilineDecorator(t *testing.T) {
	content := `from fastapi import APIRouter

router = APIRouter(prefix="/users")

@router.get(
    "/",
    response_model=UsersPublic,
)
def read_users():
    return {}
`
	routes := parseFastAPI("app/api/routes/users.py", content)

	assertRoute(t, routes, "GET", "/users", "read_users", "fastapi")
}

func TestParseFastAPIIgnoresPatchDecoratorsWithoutFastAPISignal(t *testing.T) {
	content := `from unittest.mock import patch

@patch("oscar.apps.checkout.session.CheckoutSessionMixin.skip_unless_basket_requires_shipping")
def test_check_basket_is_valid(mock_skip):
    pass

@mock.patch("oscar.apps.catalogue.abstract_models.find")
def test_symlink_creates_directories(mock_find):
    pass
`
	routes := parseFastAPI("tests/test_checkout.py", content)

	if len(routes) != 0 {
		t.Fatalf("routes = %+v, want none for non-FastAPI patch decorators", routes)
	}
}

func TestExtractFastAPIImportedRouterPrefix(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "app/main.py", `from fastapi import FastAPI
from app.api.routes.users import router as users_router

app = FastAPI()
app.include_router(users_router, prefix="/api/v1")
`)
	writeRouteFile(t, root, "app/api/routes/users.py", `from fastapi import APIRouter

router = APIRouter(prefix="/users")

@router.get("/me")
def read_me():
    return {}
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/api/v1/users/me", "read_me", "fastapi")
	assertNoRoute(t, routes, "GET", "/users/me", "read_me", "fastapi")
}

func TestExtractFastAPIComposedRouterStaticPrefix(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "app/main.py", `from fastapi import FastAPI
from app.api.main import api_router
from app.core.config import settings

app = FastAPI()
app.include_router(api_router, prefix=settings.API_V1_STR)
`)
	writeRouteFile(t, root, "app/api/main.py", `from fastapi import APIRouter
from app.api.routes import users

api_router = APIRouter()
api_router.include_router(users.router)
`)
	writeRouteFile(t, root, "app/api/routes/users.py", `from fastapi import APIRouter

router = APIRouter(prefix="/users")

@router.get("/me")
def read_me():
    return {}
`)
	writeRouteFile(t, root, "app/core/config.py", `API_V1_STR: str = "/api/v1"

class Settings:
    pass

settings = Settings()
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/api/v1/users/me", "read_me", "fastapi")
	assertNoRoute(t, routes, "GET", "/users/me", "read_me", "fastapi")
}

func TestExtractFastAPIUnresolvedImportKeepsChildRoute(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "app/main.py", `from fastapi import FastAPI

app = FastAPI()
app.include_router(users_router, prefix="/api/v1")
`)
	writeRouteFile(t, root, "app/api/routes/users.py", `from fastapi import APIRouter

router = APIRouter(prefix="/users")

@router.get("/me")
def read_me():
    return {}
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/users/me", "read_me", "fastapi")
	assertNoRoute(t, routes, "GET", "/api/v1/users/me", "read_me", "fastapi")
}

func TestParseExpressWithRouterPrefix(t *testing.T) {
	content := `const express = require("express")
const app = express()
const orderRouter = express.Router()

orderRouter.post("/create", orderController.create)
orderRouter.get("status", statusHandler)
app.use("/api/orders", orderRouter)
`
	routes := parseExpress("src/routes/order.js", content)

	assertRoute(t, routes, "POST", "/api/orders/create", "orderController.create", "express")
	assertRoute(t, routes, "GET", "/api/orders/status", "statusHandler", "express")
}

func TestParseExpressMultilineRoutes(t *testing.T) {
	content := `import { Router } from "express"

const router = Router()

router.get(
  "/articles/feed",
  auth.required,
  async (req, res) => {}
)
`
	routes := parseExpress("src/routes/article.controller.ts", content)

	assertRoute(t, routes, "GET", "/articles/feed", "auth.required", "express")
}

func TestParseExpressIgnoresFrontendHTTPClientCalls(t *testing.T) {
	content := `import agent from "./agent"

const requests = {
  get: (url) => agent.get(url),
  post: (url, body) => agent.post(url, body),
  put: (url, body) => agent.put(url, body),
}

export const Articles = {
  all: () => requests.get("/articles"),
  create: (article) => requests.post("/articles", { article }),
}

export const User = {
  current: () => requests.get("/user"),
  update: (user) => requests.put("/user", { user }),
}
`
	routes := parseExpress("src/api.js", content)

	if len(routes) != 0 {
		t.Fatalf("routes = %+v, want none for frontend HTTP client calls", routes)
	}
}

func TestExtractExpressRequireRouterPrefix(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "src/app.js", `const express = require("express")
const app = express()
const orderRouter = require("./routes/order")

app.use("/api/orders", orderRouter)
`)
	writeRouteFile(t, root, "src/routes/order.js", `const express = require("express")
const router = express.Router()

router.post("/create", orderController.create)
router.get("status", statusHandler)

module.exports = router
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "POST", "/api/orders/create", "orderController.create", "express")
	assertRoute(t, routes, "GET", "/api/orders/status", "statusHandler", "express")
	assertNoRoute(t, routes, "POST", "/create", "orderController.create", "express")
}

func TestExtractExpressImportRouterPrefix(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "src/app.ts", `import express from "express"
import userRouter from "./routes/users"

const app = express()
app.use("/api/users", userRouter)
`)
	writeRouteFile(t, root, "src/routes/users.ts", `import express from "express"
const router = express.Router()

router.get("/me", userController.me)

export default router
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/api/users/me", "userController.me", "express")
	assertNoRoute(t, routes, "GET", "/me", "userController.me", "express")
}

func TestExtractExpressComposedRouterPrefix(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "src/routes.ts", `import { Router } from "express"
import tagRouter from "./tag.controller"
import articleRouter from "./article.controller"

const api = Router()
  .use(tagRouter)
  .use(articleRouter);

export default Router().use("/api", api);
`)
	writeRouteFile(t, root, "src/tag.controller.ts", `import { Router } from "express"
const router = Router()

router.get("/tags", listTags)

export default router
`)
	writeRouteFile(t, root, "src/article.controller.ts", `import { Router } from "express"
const router = Router()

router.get("/articles", listArticles)

export default router
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "GET", "/api/tags", "listTags", "express")
	assertRoute(t, routes, "GET", "/api/articles", "listArticles", "express")
	assertNoRoute(t, routes, "GET", "/tags", "listTags", "express")
}

func TestExtractExpressUnresolvedImportKeepsChildRoute(t *testing.T) {
	root := t.TempDir()
	writeRouteFile(t, root, "src/app.js", `const express = require("express")
const app = express()

app.use("/api/orders", orderRouter)
`)
	writeRouteFile(t, root, "src/routes/order.js", `const express = require("express")
const router = express.Router()

router.post("/create", orderController.create)

module.exports = router
`)

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertRoute(t, routes, "POST", "/create", "orderController.create", "express")
	assertNoRoute(t, routes, "POST", "/api/orders/create", "orderController.create", "express")
}

func assertNoRoute(t *testing.T, routes []ir.APIRoute, method string, path string, handler string, source string) {
	t.Helper()
	for _, route := range routes {
		if route.Method == method && route.Path == path && route.Handler == handler && route.Source == source {
			t.Fatalf("unexpected route %s %s %s %s in %+v", method, path, handler, source, routes)
		}
	}
}

func assertRoute(t *testing.T, routes []ir.APIRoute, method string, path string, handler string, source string) {
	t.Helper()
	for _, route := range routes {
		if route.Method == method && route.Path == path && route.Handler == handler && route.Source == source {
			if route.Line <= 0 {
				t.Fatalf("route %s %s line = %d, want positive line", method, path, route.Line)
			}
			if route.Confidence == "" {
				t.Fatalf("route %s %s confidence is empty", method, path)
			}
			if route.Evidence == "" {
				t.Fatalf("route %s %s evidence is empty", method, path)
			}
			return
		}
	}
	t.Fatalf("missing route %s %s %s %s in %+v", method, path, handler, source, routes)
}

func writeRouteFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
