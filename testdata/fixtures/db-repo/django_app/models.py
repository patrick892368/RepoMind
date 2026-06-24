from django.db import models


class Customer(models.Model):
    id = models.BigAutoField(primary_key=True)
    email = models.EmailField(unique=True)
    nickname = models.CharField(max_length=64, null=True, blank=True)

    class Meta:
        db_table = "customers"


class Invoice(models.Model):
    customer = models.ForeignKey(Customer, on_delete=models.CASCADE)
    code = models.CharField(max_length=32, unique=True)
    total = models.DecimalField(max_digits=12, decimal_places=2)
