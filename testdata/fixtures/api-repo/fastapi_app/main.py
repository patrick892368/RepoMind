from fastapi import APIRouter, FastAPI

app = FastAPI()
router = APIRouter()


@app.post("/login")
def login():
    return {"ok": True}


@router.get("/wallet/info")
async def wallet_info():
    return {"balance": 100}
