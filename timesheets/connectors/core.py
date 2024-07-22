"""Core interfaces and dataclasses for timesheets connectors."""

from __future__ import annotations

__all__ = ["TimeEntry", "SourceConnector", "TargetConnector"]

import abc
import dataclasses
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    import datetime


@dataclasses.dataclass(frozen=True, kw_only=True)
class TimeEntry:
    """Metadata about a single time entry."""

    issue: str
    from_: datetime.datetime
    till_: datetime.datetime
    description: str | None = None
    tags: list[str] = dataclasses.field(default_factory=list)

    def __str__(self) -> str:
        """Represent TimeEntry as a string."""
        tags: str = ",".join(self.tags)

        form_: str = self.from_.isoformat()
        till_: str = self.till_.isoformat()
        description: str = self.description or "-"

        return f"[{form_} - {till_}] [{self.issue}] {description} [{tags}]"


class SourceConnector(metaclass=abc.ABCMeta):
    """Connector for getting time entries."""

    @abc.abstractmethod
    def get_time_entries(
        self,
        from_: datetime.datetime,
        till_: datetime.datetime,
    ) -> list[TimeEntry]:
        """Get time entries for a given time range."""
        raise NotImplementedError


class TargetConnector(metaclass=abc.ABCMeta):
    """Connector for logging time entries."""

    @abc.abstractmethod
    def create_time_entry(self, entry: TimeEntry) -> None:
        """Create a new time entry."""
        raise NotImplementedError
