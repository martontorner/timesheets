#!/usr/bin/env python3
# coding=utf-8
from __future__ import annotations

import dataclasses
import datetime
import json
import math
import re
import typing
import urllib.parse
from pathlib import Path

import click
import pytz
import requests
import requests.auth


TZ: pytz.tzinfo = pytz.timezone('Europe/Budapest')
NOW: datetime.datetime = datetime.datetime.now(tz=TZ)

WorkLog: typing.TypeAlias = dict[str, dict[str, typing.Union[str, int, float]]]


@dataclasses.dataclass(frozen=True, kw_only=True)
class TimeEntry(object):
    issue: str
    from_: datetime.datetime
    till_: datetime.datetime
    description: str | None = None
    tags: list[str] = dataclasses.field(default_factory=list)


class Toggl(object):
    BASE_URL: str = "https://api.track.toggl.com/api/v9"

    def __init__(self, token: str, workspace_id: str) -> None:
        self._token: str = token
        self._workspace_id: str = workspace_id

    def _create_time_entry_from_toggl_entry(
        self,
        entry: dict[str, typing.Any],
    ) -> TimeEntry:
        match = re.search(r"\[([A-Za-z\d\-]+)]", entry["description"])

        start: datetime.datetime = \
            datetime.datetime.fromisoformat(entry["start"]).astimezone(tz=TZ)
        stop: datetime.datetime = start + datetime.timedelta(seconds=entry["duration"])

        if match is not None:
            issue = match.group(1)
            description = entry["description"].replace(f"[{issue}]", "").strip(" ")
        else:
            issue = entry["description"]
            description = None

        return TimeEntry(
            issue=issue,
            from_=start,
            till_=stop,
            description=description,
            tags=entry["tags"],
        )

    def get_time_entries(
        self,
        from_: datetime.datetime,
        till_: datetime.datetime,
    ) -> list[TimeEntry]:
        entries: list[TimeEntry] = []

        from_str: str = urllib.parse.quote_plus(from_.isoformat())
        till_str: str = urllib.parse.quote_plus(till_.isoformat())

        query: str = f"start_date={from_str}&end_date={till_str}"
        url: str = f"{Toggl.BASE_URL}/me/time_entries?{query}"

        auth: requests.auth.HTTPBasicAuth = requests.auth.HTTPBasicAuth(
            username=self._token,
            password="api_token",
        )

        response: requests.Response = requests.get(url=url, auth=auth)

        if not response.ok:
            raise Exception(f"Cannot get Toggl entries ({response.status_code})")

        for entry in response.json():
            if entry["workspace_id"] == self._workspace_id and entry["stop"] is not None:
                entries.append(self._create_time_entry_from_toggl_entry(entry))

        return entries


class Kronos(object):
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

    def _create_work_log(self, work_log: WorkLog) -> None:
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

        work_log: WorkLog = {
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


@click.command()
@click.argument("path", type=click.Path(exists=True, file_okay=True, resolve_path=True, readable=True))
@click.option("--from", "from_", type=click.DateTime(), default=NOW.replace(hour=0, minute=0, second=0, microsecond=0))
@click.option("--till", "till_", type=click.DateTime(), default=NOW)
@click.option('--dry', is_flag=True, default=False)
@click.option('--stop-on-fail', is_flag=True, default=False)
@click.password_option(confirmation_prompt=False)
def main(
    path: str,
    from_: datetime.datetime,
    till_: datetime.datetime,
    dry: bool,
    stop_on_fail: bool,
    password: str,
) -> None:
    config: dict[str, typing.Any] = json.loads(s=Path(path).read_text())

    from_ = from_.astimezone(tz=TZ)
    till_ = till_.astimezone(tz=TZ)

    toggl: Toggl = Toggl(
        token=config["toggl"]["token"],
        workspace_id=config["toggl"]["workspace_id"],
    )

    kronos: Kronos = Kronos(
        username=config["kronos"]["username"],
        password=password,
        tags=config["kronos"]["tags"],
    )

    for entry in toggl.get_time_entries(from_=from_, till_=till_):
        if not dry:
            try:
                kronos.create_time_entry(entry=entry)
                print(f"LOGGED: {entry}")
            except Exception as e:
                if stop_on_fail:
                    print(f"FAILED: {entry}, reason: {e}")
                    raise e from None
                else:
                    print(f"IGNORE: {entry}, reason: {e}")
        else:
            print(f"PARSED: {entry}")


if __name__ == '__main__':
    main()
