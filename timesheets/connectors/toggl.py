# coding=utf-8
__all__ = ["Toggl"]

import datetime
import re
import urllib.parse

import requests
import requests.auth
import typing

from timesheets.connectors.core import TimeEntry, SourceConnector
from timesheets.constants import TZ


class Toggl(SourceConnector):
    BASE_URL: str = "https://api.track.toggl.com/api/v9"

    def __init__(self, token: str, workspace_id: str) -> None:
        self._token: str = token
        self._workspace_id: str = workspace_id

    # noinspection PyMethodMayBeStatic
    def _create_time_entry_from_toggl_entry(
        self,
        entry: dict[str, typing.Any],
    ) -> TimeEntry:
        match: re.Match = re.search(r"\[([A-Za-z\d\-]+)]", entry["description"])

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
