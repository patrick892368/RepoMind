from fastapi import FastAPI

app = FastAPI()


@app.post("/order/create")
def create_order():
    return {"ok": True}
