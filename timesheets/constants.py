"""Global constants for timesheets."""

import datetime

import pytz

TZ: pytz.tzinfo = pytz.timezone("Europe/Budapest")
NOW: datetime.datetime = datetime.datetime.now(tz=TZ)

TODAY: datetime.datetime = NOW.replace(
    hour=0,
    minute=0,
    second=0,
    microsecond=0,
)
