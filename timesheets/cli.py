# coding=utf-8
from __future__ import annotations

__all__ = ["cli"]

import datetime
import json
import pathlib

import click

from timesheets.connectors.kronos import Kronos
from timesheets.connectors.toggl import Toggl
from timesheets.constants import NOW, TZ


@click.group()
@click.option("--config", "path", default=pathlib.Path.home() / ".config" / "timesheets" / "config.json")
@click.version_option()
@click.pass_context
def cli(ctx: click.Context, path: pathlib.Path | str) -> None:
    ctx.ensure_object(dict)

    path = pathlib.Path(path)

    if not path.is_file():
        raise FileNotFoundError(f"The file [{path}] does not exist")

    ctx.obj["config"] = json.loads(s=path.read_text())


@cli.command()
@click.pass_context
def config(ctx: click.Context) -> None:
    print(json.dumps(obj=ctx.obj["config"], indent=4))


@cli.command()
@click.option("--from", "from_", type=click.DateTime(), default=NOW.replace(hour=0, minute=0, second=0, microsecond=0))
@click.option("--till", "till_", type=click.DateTime(), default=NOW)
@click.option('--dry', is_flag=True, default=False)
@click.option('--stop-on-fail', is_flag=True, default=False)
@click.password_option(confirmation_prompt=False)
@click.pass_context
def sync(
    ctx: click.Context,
    from_: datetime.datetime,
    till_: datetime.datetime,
    dry: bool,
    stop_on_fail: bool,
    password: str,
) -> None:

    from_ = from_.astimezone(tz=TZ)
    till_ = till_.astimezone(tz=TZ)

    toggl: Toggl = Toggl(
        token=ctx.obj["config"]["toggl"]["token"],
        workspace_id=ctx.obj["config"]["toggl"]["workspace_id"],
    )

    kronos: Kronos = Kronos(
        username=ctx.obj["config"]["kronos"]["username"],
        password=password,
        tags=ctx.obj["config"]["kronos"]["tags"],
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
