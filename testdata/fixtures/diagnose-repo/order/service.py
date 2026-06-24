from app.cache import cache
from app.tasks import send_order_notify


def update_order_status(order):
    order.status = "paid"
    order.save()
    cache.set(f"order:{order.id}:status", order.status)
    send_order_notify.delay(order.id)


def update_wallet(wallet, amount):
    wallet.balance = wallet.balance - amount
    wallet.save()
