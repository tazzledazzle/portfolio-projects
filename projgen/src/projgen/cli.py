#!/usr/bin/env python3
"""
projgen: A CLI tool to scaffold Bazel/Gradle projects with opinionated defaults.
"""
import sys
import click

from .config import DEFAULTS
from .interactive import interactive_setup, prompt_for_missing_args
from .validation import TemplateValidator
from .plugins import PluginManager
from generators.scaffold import scaffold_project
from .telemetry import init_telemetry, record_event


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
    ctx.obj["plugin_manager"] = PluginManager()
    if ctx.obj["telemetry_enabled"]:
        init_telemetry()


@cli.command()
@click.argument("project_name", required=False)
@click.option(
    "--build",
    type=click.Choice(["bazel", "gradle", "both"]),
    default=None,
    help="Build system to scaffold",
)
@click.option(
    "--language",
    type=click.Choice(DEFAULTS["languages"]),
    default=None,
    help="Primary project language",
)
@click.option(
    "--license",
    default=None,
    help="License identifier (e.g., MIT, Apache-2.0)",
)
@click.option(
    "--ci",
    type=click.Choice(DEFAULTS["ci_providers"] + ["none"]),
    default=None,
    help="CI provider to configure",
)
@click.option(
    "--overwrite",
    is_flag=True,
    default=False,
    help="Overwrite existing files if present",
)
@click.option(
    "--interactive", "-i",
    is_flag=True,
    default=False,
    help="Run interactive setup",
)
@click.option(
    "--validate-only",
    is_flag=True,
    default=False,
    help="Only validate configuration, don't generate",
)
@click.pass_context
def init(ctx, project_name, build, language, license, ci, overwrite, interactive, validate_only):
    """Initialize a new project scaffold"""
    plugin_manager = ctx.obj["plugin_manager"]
    
    # Interactive mode
    if interactive or not project_name:
        config = interactive_setup()
        if not config:
            return
    else:
        # Prompt for missing required arguments
        config = prompt_for_missing_args(
            project_name=project_name,
            build_system=build or DEFAULTS["build"],
            language=language,
            license_id=license or DEFAULTS["license"],
            ci_provider=ci or "github",
            overwrite=overwrite,
            features={"docker": False, "docs": False, "pre_commit": True}
        )
    
    # Validate configuration
    validator = TemplateValidator(DEFAULTS["templates_dir"])
    if not validator.validate_project_config(config):
        click.echo("❌ Configuration validation failed:")
        click.echo(validator.get_validation_report())
        if not click.confirm("Continue anyway?", default=False):
            return
    
    if validate_only:
        click.echo("✅ Configuration is valid!")
        click.echo(validator.get_validation_report())
        return
    
    # Record telemetry event
    if ctx.obj.get("telemetry_enabled"):
        record_event("init_command", config)

    click.echo(f"🚀 Scaffolding '{config['project_name']}'...")
    
    try:
        # Generate basic project structure
        scaffold_project(
            project_name=config["project_name"],
            build_system=config["build_system"],
            language=config["language"],
            license_id=config["license_id"],
            ci_provider=config["ci_provider"],
            overwrite=config["overwrite"],
        )
        
        # Apply plugins
        from pathlib import Path
        project_path = Path(config["project_name"])
        plugin_manager.apply_plugins(project_path, config)
        
        click.echo("✅ Project scaffold generated successfully!")
        
        # Show next steps
        _show_next_steps(config)
        
    except Exception as e:
        click.echo(f"❌ Error: {e}", err=True)
        sys.exit(1)


if __name__ == "__main__":
    cli(obj={})

def _show_next_steps(config: dict):
    """Show next steps after project generation."""
    project_name = config["project_name"]
    
    click.echo("\n📋 Next Steps:")
    click.echo(f"  1. cd {project_name}")
    
    if config.get("features", {}).get("pre_commit"):
        click.echo("  2. pre-commit install")
    
    if config["build_system"] in ["gradle", "both"]:
        click.echo("  3. ./gradlew build")
    
    if config["build_system"] in ["bazel", "both"]:
        click.echo("  4. bazel build //...")
    
    if config.get("features", {}).get("docker"):
        click.echo("  5. docker-compose up --build")
    
    click.echo("  6. Start coding! 🎉")


@cli.command()
@click.pass_context
def plugins(ctx):
    """List available plugins"""
    plugin_manager = ctx.obj["plugin_manager"]
    click.echo(plugin_manager.get_plugin_info())


@cli.command()
@click.argument("templates_dir", required=False)
@click.pass_context
def validate(ctx, templates_dir):
    """Validate templates and configuration"""
    templates_path = templates_dir or DEFAULTS["templates_dir"]
    
    validator = TemplateValidator(templates_path)
    
    click.echo("🔍 Validating templates...")
    is_valid = validator.validate_templates()
    
    click.echo(validator.get_validation_report())
    
    if is_valid:
        click.echo("✅ All templates are valid!")
    else:
        click.echo("❌ Template validation failed!")
        sys.exit(1)


@cli.command()
@click.argument("plugin_path")
@click.pass_context
def install_plugin(ctx, plugin_path):
    """Install a plugin from a file"""
    plugin_manager = ctx.obj["plugin_manager"]
    
    plugin = plugin_manager.load_plugin_from_file(plugin_path)
    if plugin:
        click.echo(f"✅ Plugin '{plugin.name}' installed successfully!")
    else:
        click.echo(f"❌ Failed to install plugin from {plugin_path}")
        sys.exit(1)