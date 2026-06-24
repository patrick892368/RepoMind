from sqlalchemy import ForeignKey, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship


class Base(DeclarativeBase):
    pass


class Team(Base):
    __tablename__ = "teams"

    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(255), unique=True)
    users: Mapped[list["MappedUser"]] = relationship(back_populates="team")


class MappedUser(Base):
    __tablename__ = "mapped_users"

    id: Mapped[int] = mapped_column(primary_key=True)
    email: Mapped[str | None] = mapped_column(String, unique=True, nullable=True)
    team_id: Mapped[int] = mapped_column(ForeignKey("teams.id"))
    team: Mapped["Team"] = relationship(back_populates="users")
