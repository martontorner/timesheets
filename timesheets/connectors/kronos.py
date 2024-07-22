"""Kronos connector for timesheets."""

from __future__ import annotations

__all__ = ["Kronos"]

import math
import typing

import requests

from timesheets.connectors.core import TargetConnector, TimeEntry

type _WorkLog = dict[str, dict[str, str | int | float]]


class Kronos(TargetConnector):
    """Target connector for Kronos timesheet logger via the JIRA API."""

    BASE_URL: str = "https://jira.capsys.hu"
    DEFAULT_COMMENT: str = ""
    DEFAULT_SITE_ID: int = 31
    DEFAULT_TIMEOUT_S: int = 10

    def __init__(
        self,
        username: str,
        password: str,
        tags: dict[str, dict[str, typing.Any]],
    ) -> None:
        """Initialize Kronos connector."""
        self._username: str = username
        self._password: str = password

        self._tags: dict[str, dict[str, typing.Any]] = tags

        self._headers: dict[str, str] | None = None

        self.login()

    def _is_logged_in(self) -> bool:
        return self._headers is not None

    def login(self) -> None:
        """Login to Kronos API and create reusable session."""
        url: str = f"{Kronos.BASE_URL}/rest/auth/1/session"

        response: requests.Response = requests.post(
            url=url,
            json={"username": self._username, "password": self._password},
            timeout=Kronos.DEFAULT_TIMEOUT_S,
        )

        if not response.ok:
            msg: str = f"Cannot login to JIRA ({response.status_code})"
            raise Exception(msg)  # noqa: TRY002

        result: dict[str, dict[str, str]] = response.json()
        name: str = result["session"]["name"]
        value: str = result["session"]["value"]

        self._headers = {"cookie": f"{name}={value}"}

    def _ensure_login(self) -> None:
        if not self._is_logged_in():
            self.login()

    def _is_valid_issue(self, issue: str) -> bool:
        url: str = f"{Kronos.BASE_URL}/rest/api/latest/issue/{issue}"

        self._ensure_login()

        response: requests.Response = requests.get(
            url=url,
            headers=self._headers,
            timeout=Kronos.DEFAULT_TIMEOUT_S,
        )

        return response.ok

    def _create_work_log(self, work_log: _WorkLog) -> None:
        url: str = f"{Kronos.BASE_URL}/rest/kronos/1.0/log-entry"
        issue: str = work_log["worklogInput"]["issueKey"]

        self._ensure_login()

        if not self._is_valid_issue(issue=issue):
            msg: str = f"Unknown JIRA issue [{issue}]"
            raise Exception(msg)  # noqa: TRY002

        response: requests.Response = requests.post(
            url=url,
            json=work_log,
            headers=self._headers,
            timeout=Kronos.DEFAULT_TIMEOUT_S,
        )

        if not response.ok:
            status_code: int = response.status_code
            reason: str = response.text

            msg: str = (
                f"Cannot create work log ({status_code}, reason: {reason})"
            )
            raise Exception(msg)  # noqa: TRY002

    def create_time_entry(self, entry: TimeEntry) -> None:
        """Create new time entry in Kronos."""
        time_spent: int = math.floor(
            (entry.till_ - entry.from_).total_seconds() / 60,
        )

        issue_key: str = entry.issue
        start_offset_date_time: str = entry.from_.replace(
            second=0,
            microsecond=0,
        ).isoformat()
        comment: str = entry.description or Kronos.DEFAULT_COMMENT
        site_id: int = Kronos.DEFAULT_SITE_ID

        work_log: _WorkLog = {
            "worklogInput": {
                "issueKey": issue_key,
                "timeSpent": time_spent,
                "startOffsetDateTime": start_offset_date_time,
                "comment": comment,
                "siteId": site_id,
            },
            "travelInput": {
                "travelToTimeSpentInMinutes": 0,
                "travelFromTimeSpentInMinutes": 0,
                "fromSiteId": None,
            },
        }

        for tag in entry.tags:
            if tag in self._tags:
                work_log["worklogInput"].update(self._tags[tag])

        self._create_work_log(work_log=work_log)
