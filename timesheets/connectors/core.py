# coding=utf-8
from __future__ import annotations

__all__ = ["TimeEntry", "SourceConnector", "TargetConnector"]

import abc
import dataclasses
import datetime


@dataclasses.dataclass(frozen=True, kw_only=True)
class TimeEntry(object):
    issue: str
    from_: datetime.datetime
    till_: datetime.datetime
    description: str | None = None
    tags: list[str] = dataclasses.field(default_factory=list)

    def __str__(self) -> str:
        tags: str = ','.join(self.tags)

        return f"[{self.from_.isoformat()} - {self.till_.isoformat()}] [{self.issue}] {self.description or '-'} [{tags}]"


class SourceConnector(metaclass=abc.ABCMeta):
    @abc.abstractmethod
    def get_time_entries(
        self,
        from_: datetime.datetime,
        till_: datetime.datetime,
    ) -> list[TimeEntry]:
        raise NotImplementedError


class TargetConnector(metaclass=abc.ABCMeta):
    @abc.abstractmethod
    def create_time_entry(self, entry: TimeEntry) -> None:
        raise NotImplementedError
