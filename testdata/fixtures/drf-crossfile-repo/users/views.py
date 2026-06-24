from rest_framework.decorators import action
from rest_framework.viewsets import ModelViewSet


class UserViewSet(ModelViewSet):
    @action(detail=True, methods=["post"], url_path="set-password")
    def set_password(self, request, pk=None):
        pass
