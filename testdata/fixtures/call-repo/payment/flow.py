def pay_callback(data):
    order = update_order(data)
    update_balance(order)
    send_notify(order)
    write_log(order)


def update_order(data):
    validate_order(data)
    return data


def update_balance(order):
    write_log(order)


def send_notify(order):
    write_log(order)


def validate_order(data):
    return data


def write_log(value):
    return value
