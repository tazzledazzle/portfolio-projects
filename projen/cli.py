#!/usr/bin/env python3
"""
projgen: A CLI tool to scaffold Bazel/Gradle projects with opinionated defaults.
"""
import sys
import click

from config import DEFAULTS
from generators import scaffold_project
from telemetry import init_telemetry, record_event


@click.group()
@click.option(
    "--no-telemetry", is_flag=True, default=False, help="Disable usage reporting"
)
@click.pass_context
def cli(ctx, no_telemetry):
    """Entry point for projgen CLI"""
    # Initialize telemetry if enabled
    ctx.ensure_object(dict)
    ctx.obj["telemetry_enabled"] = not no_telemetry
    if ctx.obj["telemetry_enabled"]:
        init_telemetry()


@cli.command()
@click.argument("project_name")
@click.option(
    "--build",
    type=click.Choice(["bazel", "gradle", "both"]),
    default=DEFAULTS["build"],
    help="Build system to scaffold",
)
@click.option(
    "--language",
    type=click.Choice(DEFAULTS["languages"]),
    required=True,
    help="Primary project language",
)
@click.option(
    "--license",
    default=DEFAULTS["license"],
    help="License identifier (e.g., MIT, Apache-2.0)",
)
@click.option(
    "--ci",
    type=click.Choice(DEFAULTS["ci_providers"]),
    default=DEFAULTS["ci"],
    help="CI provider to configure",
)
@click.option(
    "--overwrite",
    is_flag=True,
    default=False,
    help="Overwrite existing files if present",
)
@click.pass_context
def init(ctx, project_name, build, language, license, ci, overwrite):
    """Initialize a new project scaffold"""
    # Record telemetry event
    if ctx.obj.get("telemetry_enabled"):
        record_event(
            "init_command",
            {"build": build, "language": language, "license": license, "ci": ci},
        )

    click.echo(
        f"Scaffolding '{project_name}' with build={build}, language={language}, license={license}, ci={ci}\nOverwrite: {overwrite}"
    )
    try:
        scaffold_project(
            project_name=project_name,
            build_system=build,
            language=language,
            license_id=license,
            ci_provider=ci,
            overwrite=overwrite,
        )
        click.echo("✅ Project scaffold generated successfully.")
    except Exception as e:
        click.echo(f"❌ Error: {e}", err=True)
        sys.exit(1)


if __name__ == "__main__":
    cli(obj={})
