# coding=utf-8
import datetime

import pytz

TZ: pytz.tzinfo = pytz.timezone('Europe/Budapest')
NOW: datetime.datetime = datetime.datetime.now(tz=TZ)
