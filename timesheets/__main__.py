"""Main entrypoint for timesheets package scripts."""

from timesheets.cli import cli


def main() -> None:
    """Call main CLI entrypoint."""
    cli()


if __name__ == "__main__":
    main()
