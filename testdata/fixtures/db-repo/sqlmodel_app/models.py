from pydantic import BaseModel
from pydantic_settings import BaseSettings
from sqlmodel import Field, Relationship, SQLModel


class Settings(BaseSettings):
    database_url: str


class Payload(BaseModel):
    value: str


class UserBase(SQLModel):
    email: str = Field(unique=True)


class UserCreate(UserBase):
    password: str


class User(UserBase, table=True):
    id: int = Field(primary_key=True)
    orders: list["SQLModelOrder"] = Relationship(back_populates="user")


class SQLModelOrder(SQLModel, table=True):
    id: int = Field(primary_key=True)
    user_id: int = Field(foreign_key="user.id")
    user: User | None = Relationship(back_populates="orders")
