"""Project scaffolding functionality."""

import os
import shutil
from pathlib import Path


def scaffold_project(project_name, build_system="both", language="python", 
                    license_id="MIT", ci_provider="github", overwrite=False):
    """
    Scaffold a new project with the specified configuration.
    
    Args:
        project_name: Name/path of the project to create
        build_system: Build system to use ("gradle", "bazel", or "both")
        language: Programming language
        license_id: License identifier
        ci_provider: CI provider
        overwrite: Whether to overwrite existing directory
    """
    project_path = Path(project_name)
    
    # Check if directory exists
    if project_path.exists() and not overwrite:
        raise FileExistsError(f"Directory {project_name} already exists")
    
    # Create project directory
    if project_path.exists() and overwrite:
        shutil.rmtree(project_path)
    
    project_path.mkdir(parents=True, exist_ok=True)
    
    # Create basic structure
    _create_common_files(project_path, project_name, license_id)
    _create_source_structure(project_path, language)
    
    if build_system in ["gradle", "both"]:
        _create_gradle_files(project_path, project_name)
    
    if build_system in ["bazel", "both"]:
        _create_bazel_files(project_path, project_name)
    
    if ci_provider == "github":
        _create_github_ci(project_path, project_name)


def _create_common_files(project_path, project_name, license_id):
    """Create common project files."""
    # README.md
    (project_path / "README.md").write_text(f"# {project_name}\n\nProject description here.\n")
    
    # LICENSE
    (project_path / "LICENSE").write_text(f"{license_id} License\n\nCopyright (c) 2025\n")
    
    # .gitignore
    (project_path / ".gitignore").write_text("*.pyc\n__pycache__/\n.DS_Store\n")


def _create_source_structure(project_path, language):
    """Create source directory structure."""
    src_main = project_path / "src" / "main" / language
    src_test = project_path / "src" / "test" / language
    
    src_main.mkdir(parents=True, exist_ok=True)
    src_test.mkdir(parents=True, exist_ok=True)


def _create_gradle_files(project_path, project_name):
    """Create Gradle build files."""
    # build.gradle
    (project_path / "build.gradle").write_text(f"// Gradle build for {project_name}\n")
    
    # settings.gradle
    (project_path / "settings.gradle").write_text(f"rootProject.name = '{project_name}'\n")


def _create_bazel_files(project_path, project_name):
    """Create Bazel build files."""
    # WORKSPACE
    (project_path / "WORKSPACE").write_text(f"# Workspace for {project_name}\n")
    
    # BUILD.bazel
    (project_path / "BUILD.bazel").write_text(f"# Build for {project_name}\n")


def _create_github_ci(project_path, project_name):
    """Create GitHub Actions CI workflow."""
    workflows_dir = project_path / ".github" / "workflows"
    workflows_dir.mkdir(parents=True, exist_ok=True)
    
    ci_content = f"# CI for {project_name}\nname: CI\n\non: [push, pull_request]\n"
    (workflows_dir / "ci.yml").write_text(ci_content)