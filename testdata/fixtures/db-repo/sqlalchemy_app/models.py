from sqlalchemy import Column, ForeignKey, Integer, String
from sqlalchemy.orm import relationship

from .database import Base


class Account(Base):
    __tablename__ = "accounts"

    id = Column(Integer, primary_key=True)
    email = Column(String, unique=True)
    profile = relationship("Profile")


class Profile(Base):
    __tablename__ = "profiles"

    id = Column(Integer, primary_key=True)
    account_id = Column(Integer, ForeignKey("accounts.id"))
    display_name = Column(String, nullable=True)
