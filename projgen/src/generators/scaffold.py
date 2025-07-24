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
    # README.md with better template
    readme_content = f"""# {project_name}

## Description

Add your project description here.

## Getting Started

### Prerequisites

- List your prerequisites here

### Installation

```bash
# Add installation instructions
```

### Usage

```bash
# Add usage examples
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the {license_id} License - see the [LICENSE](LICENSE) file for details.
"""
    (project_path / "README.md").write_text(readme_content)
    
    # Enhanced LICENSE
    license_content = _get_license_text(license_id)
    (project_path / "LICENSE").write_text(license_content)
    
    # Enhanced .gitignore
    gitignore_content = _get_gitignore_content()
    (project_path / ".gitignore").write_text(gitignore_content)


def _get_license_text(license_id: str) -> str:
    """Get full license text."""
    if license_id == "MIT":
        return """MIT License

Copyright (c) 2025

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
"""
    elif license_id == "Apache-2.0":
        return """Apache License
Version 2.0, January 2004
http://www.apache.org/licenses/

Copyright 2025

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""
    else:
        return f"{license_id} License\n\nCopyright (c) 2025\n"


def _get_gitignore_content() -> str:
    """Get comprehensive .gitignore content."""
    return """# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Dependencies
node_modules/
.pnp
.pnp.js

# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg
MANIFEST
.env
.venv
env/
venv/
ENV/
env.bak/
venv.bak/

# Java
*.class
*.jar
*.war
*.ear
*.zip
*.tar.gz
*.rar
target/

# Gradle
.gradle
build/
!gradle/wrapper/gradle-wrapper.jar

# Bazel
bazel-*

# Testing
.coverage
.pytest_cache/
.tox/
.nox/
coverage.xml
*.cover
.hypothesis/

# Temporary files
*.tmp
*.temp
"""


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