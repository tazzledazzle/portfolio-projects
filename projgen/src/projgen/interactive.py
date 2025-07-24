"""Interactive prompts for projgen CLI."""

import click
from typing import Dict, Any
from .config import DEFAULTS


def interactive_setup() -> Dict[str, Any]:
    """
    Run interactive setup to gather project configuration.
    
    Returns:
        Dictionary with project configuration
    """
    click.echo("🚀 Welcome to projgen interactive setup!")
    click.echo("Let's configure your new project step by step.\n")
    
    # Project name
    project_name = click.prompt(
        "📁 Project name",
        type=str
    )
    
    # Build system
    click.echo("\n🔧 Build System Options:")
    for i, option in enumerate(["bazel", "gradle", "both"], 1):
        click.echo(f"  {i}. {option}")
    
    build_choice = click.prompt(
        "Choose build system",
        type=click.IntRange(1, 3),
        default=3
    )
    build_system = ["bazel", "gradle", "both"][build_choice - 1]
    
    # Language
    click.echo("\n💻 Programming Language:")
    languages = DEFAULTS["languages"]
    for i, lang in enumerate(languages, 1):
        click.echo(f"  {i:2d}. {lang}")
    
    lang_choice = click.prompt(
        "Choose primary language",
        type=click.IntRange(1, len(languages)),
        default=6  # Python
    )
    language = languages[lang_choice - 1]
    
    # License
    common_licenses = ["MIT", "Apache-2.0", "GPL-3.0", "BSD-3-Clause", "Unlicense"]
    click.echo("\n📄 License:")
    for i, license_type in enumerate(common_licenses, 1):
        click.echo(f"  {i}. {license_type}")
    click.echo(f"  {len(common_licenses) + 1}. Custom")
    
    license_choice = click.prompt(
        "Choose license",
        type=click.IntRange(1, len(common_licenses) + 1),
        default=1
    )
    
    if license_choice <= len(common_licenses):
        license_id = common_licenses[license_choice - 1]
    else:
        license_id = click.prompt("Enter custom license identifier", type=str)
    
    # CI Provider
    ci_provider = click.confirm(
        "\n🔄 Set up GitHub Actions CI?",
        default=True
    )
    ci = "github" if ci_provider else "none"
    
    # Additional features
    click.echo("\n✨ Additional Features:")
    
    add_docker = click.confirm("Add Dockerfile?", default=False)
    add_docs = click.confirm("Add documentation setup (MkDocs)?", default=False)
    add_pre_commit = click.confirm("Add pre-commit hooks?", default=True)
    
    # Overwrite confirmation
    overwrite = click.confirm(
        "\n⚠️  Overwrite existing files if they exist?",
        default=False
    )
    
    config = {
        "project_name": project_name,
        "build_system": build_system,
        "language": language,
        "license_id": license_id,
        "ci_provider": ci,
        "overwrite": overwrite,
        "features": {
            "docker": add_docker,
            "docs": add_docs,
            "pre_commit": add_pre_commit
        }
    }
    
    # Summary
    click.echo("\n📋 Configuration Summary:")
    click.echo(f"  Project: {project_name}")
    click.echo(f"  Build: {build_system}")
    click.echo(f"  Language: {language}")
    click.echo(f"  License: {license_id}")
    click.echo(f"  CI: {ci}")
    click.echo(f"  Features: {', '.join([k for k, v in config['features'].items() if v])}")
    
    if not click.confirm("\n✅ Proceed with this configuration?", default=True):
        click.echo("❌ Setup cancelled.")
        return None
    
    return config


def prompt_for_missing_args(**kwargs) -> Dict[str, Any]:
    """
    Prompt for any missing required arguments.
    
    Args:
        **kwargs: Current argument values
        
    Returns:
        Updated configuration dictionary
    """
    config = dict(kwargs)
    
    if not config.get("language"):
        languages = DEFAULTS["languages"]
        click.echo("Available languages:")
        for i, lang in enumerate(languages, 1):
            click.echo(f"  {i:2d}. {lang}")
        
        choice = click.prompt(
            "Choose language",
            type=click.IntRange(1, len(languages))
        )
        config["language"] = languages[choice - 1]
    
    return config