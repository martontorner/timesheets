# coding=utf-8
from __future__ import annotations

__all__ = ["Kronos"]

import math
import typing

import requests

from .core import TimeEntry, TargetConnector

_WorkLog: typing.TypeAlias = dict[str, dict[str, typing.Union[str, int, float]]]


class Kronos(TargetConnector):
    BASE_URL: str = "https://jira.capsys.hu/rest/kronos/1.0"
    DEFAULT_COMMENT: str = ""
    DEFAULT_SITE_ID: int = 31

    def __init__(
        self,
        username: str,
        password: str,
        tags: dict[str, dict[str, typing.Any]],
    ) -> None:
        self._username: str = username
        self._password: str = password

        self._tags: dict[str, dict[str, typing.Any]] = tags

        self._headers: dict[str, str] | None = None

        self.login()

    def _is_logged_in(self) -> bool:
        return self._headers is not None

    def login(self) -> None:
        url: str = "https://jira.capsys.hu/rest/auth/1/session"

        response: requests.Response = requests.post(
            url=url,
            json={"username": self._username, "password": self._password},
        )

        if not response.ok:
            raise Exception(f"Cannot login to JIRA ({response.status_code})")

        result: dict[str, dict[str, str]] = response.json()
        name: str = result['session']['name']
        value: str = result['session']['value']

        self._headers = {"cookie": f"{name}={value}"}

    def _ensure_login(self) -> None:
        if not self._is_logged_in():
            self.login()

    def _is_valid_issue(self, issue: str) -> bool:
        url: str = f"https://jira.capsys.hu/rest/api/latest/issue/{issue}"

        self._ensure_login()

        return requests.get(url=url, headers=self._headers).ok

    def _create_work_log(self, work_log: _WorkLog) -> None:
        url: str = f"{Kronos.BASE_URL}/log-entry"
        issue: str = work_log["worklogInput"]["issueKey"]

        self._ensure_login()

        if not self._is_valid_issue(issue=issue):
            raise Exception(f"Unknown JIRA issue [{issue}]")

        response: requests.Response = requests.post(
            url=url,
            json=work_log,
            headers=self._headers,
        )

        if not response.ok:
            raise Exception(f"Cannot create work log ({response.status_code})")

    def create_time_entry(self, entry: TimeEntry) -> None:
        time_spent: int = math.floor((entry.till_ - entry.from_).total_seconds() / 60)

        work_log: _WorkLog = {
            "worklogInput": {
                "issueKey": entry.issue,
                "timeSpent": time_spent,
                "startOffsetDateTime": entry.from_.isoformat(),
                "comment": entry.description or Kronos.DEFAULT_COMMENT,
                "siteId": Kronos.DEFAULT_SITE_ID,
            },
            "travelInput": {
                "travelToTimeSpentInMinutes": 0,
                "travelFromTimeSpentInMinutes": 0,
                "fromSiteId": None,
            }
        }

        for tag in entry.tags:
            if tag in self._tags:
                work_log["worklogInput"].update(self._tags[tag])

        self._create_work_log(work_log=work_log)
