from django.urls import path

from . import views

urlpatterns = [
    path("login/", views.login_view, name="login"),
    path("order/create/", views.create_order, name="create_order"),
]
