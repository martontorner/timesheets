"""CLI endpoints for timesheets."""

from __future__ import annotations

__all__ = ["cli"]

import json
import logging
import pathlib
from typing import TYPE_CHECKING

import click

from timesheets.connectors.kronos import Kronos
from timesheets.connectors.toggl import Toggl
from timesheets.constants import NOW, TZ

if TYPE_CHECKING:
    import datetime


@click.group()
@click.option(
    "--config",
    "path",
    default=pathlib.Path.home() / ".config" / "timesheets" / "config.json",
)
@click.version_option()
@click.pass_context
def cli(ctx: click.Context, path: pathlib.Path | str) -> None:
    """Call the default CLI entrypoint."""
    ctx.ensure_object(dict)

    path = pathlib.Path(path)

    if not path.is_file():
        msg = f"The file [{path}] does not exist"
        raise FileNotFoundError(msg)

    ctx.obj["config"] = json.loads(s=path.read_text())


@cli.command()
@click.pass_context
def config(ctx: click.Context) -> None:
    """Print parsed configuration."""
    print(json.dumps(obj=ctx.obj["config"], indent=4))  # noqa: T201


@cli.command()
@click.option(
    "--from",
    "from_",
    type=click.DateTime(),
    default=NOW.replace(hour=0, minute=0, second=0, microsecond=0),
)
@click.option("--till", "till_", type=click.DateTime(), default=NOW)
@click.option("--dry", is_flag=True, default=False)
@click.option("--stop-on-fail", is_flag=True, default=False)
@click.password_option(confirmation_prompt=False)
@click.pass_context
def sync(  # noqa: PLR0913
    ctx: click.Context,
    from_: datetime.datetime,
    till_: datetime.datetime,
    *,
    dry: bool,
    stop_on_fail: bool,
    password: str,
) -> None:
    """Synchronize time entries."""
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

                msg: str = f"LOGGED: {entry}"
                logging.info(msg)
            except Exception as e:
                if stop_on_fail:
                    msg: str = f"FAILED: {entry}"
                    logging.exception(msg, exc_info=e)
                else:
                    msg: str = f"IGNORE: {entry}"
                    logging.warning(msg, exc_info=e)
        else:
            msg: str = f"PARSED: {entry}"
            logging.info(msg)
