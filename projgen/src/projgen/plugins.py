"""Plugin system for projgen."""

import os
import sys
import importlib
import importlib.util
from pathlib import Path
from typing import Dict, List, Any, Optional, Callable
from abc import ABC, abstractmethod
import yaml


class ProjectPlugin(ABC):
    """Base class for projgen plugins."""
    
    @property
    @abstractmethod
    def name(self) -> str:
        """Plugin name."""
        pass
    
    @property
    @abstractmethod
    def version(self) -> str:
        """Plugin version."""
        pass
    
    @property
    @abstractmethod
    def description(self) -> str:
        """Plugin description."""
        pass
    
    @abstractmethod
    def configure(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """
        Configure the plugin with project settings.
        
        Args:
            config: Project configuration
            
        Returns:
            Updated configuration
        """
        pass
    
    @abstractmethod
    def generate_files(self, project_path: Path, config: Dict[str, Any]) -> None:
        """
        Generate files for this plugin.
        
        Args:
            project_path: Path to the project directory
            config: Project configuration
        """
        pass
    
    def post_generate(self, project_path: Path, config: Dict[str, Any]) -> None:
        """
        Optional post-generation hook.
        
        Args:
            project_path: Path to the project directory
            config: Project configuration
        """
        pass


class DockerPlugin(ProjectPlugin):
    """Plugin for Docker support."""
    
    @property
    def name(self) -> str:
        return "docker"
    
    @property
    def version(self) -> str:
        return "1.0.0"
    
    @property
    def description(self) -> str:
        return "Adds Docker support with Dockerfile and docker-compose.yml"
    
    def configure(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Configure Docker plugin."""
        if config.get("features", {}).get("docker", False):
            config.setdefault("docker", {})
            config["docker"]["base_image"] = self._get_base_image(config.get("language"))
            config["docker"]["port"] = self._get_default_port(config.get("language"))
        return config
    
    def _get_base_image(self, language: str) -> str:
        """Get appropriate base image for language."""
        base_images = {
            "python": "python:3.12-slim",
            "java": "openjdk:17-jre-slim",
            "kotlin": "openjdk:17-jre-slim",
            "node": "node:18-alpine",
            "typescript": "node:18-alpine",
            "rust": "rust:1.70-slim",
            "cpp": "gcc:latest",
            "c": "gcc:latest"
        }
        return base_images.get(language, "ubuntu:22.04")
    
    def _get_default_port(self, language: str) -> int:
        """Get default port for language."""
        default_ports = {
            "python": 8000,
            "java": 8080,
            "kotlin": 8080,
            "node": 3000,
            "typescript": 3000,
            "rust": 8000,
            "cpp": 8080,
            "c": 8080
        }
        return default_ports.get(language, 8080)
    
    def generate_files(self, project_path: Path, config: Dict[str, Any]) -> None:
        """Generate Docker files."""
        if not config.get("features", {}).get("docker", False):
            return
        
        language = config.get("language", "python")
        docker_config = config.get("docker", {})
        
        # Generate Dockerfile
        dockerfile_content = self._generate_dockerfile(language, docker_config)
        (project_path / "Dockerfile").write_text(dockerfile_content)
        
        # Generate docker-compose.yml
        compose_content = self._generate_docker_compose(config)
        (project_path / "docker-compose.yml").write_text(compose_content)
        
        # Generate .dockerignore
        dockerignore_content = self._generate_dockerignore(language)
        (project_path / ".dockerignore").write_text(dockerignore_content)
    
    def _generate_dockerfile(self, language: str, docker_config: Dict[str, Any]) -> str:
        """Generate Dockerfile content."""
        base_image = docker_config.get("base_image", "ubuntu:22.04")
        port = docker_config.get("port", 8080)
        
        if language == "python":
            return f"""FROM {base_image}

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE {port}

CMD ["python", "src/main.py"]
"""
        elif language in ["java", "kotlin"]:
            return f"""FROM {base_image}

WORKDIR /app

COPY build/libs/*.jar app.jar

EXPOSE {port}

CMD ["java", "-jar", "app.jar"]
"""
        elif language in ["node", "typescript"]:
            return f"""FROM {base_image}

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE {port}

CMD ["npm", "start"]
"""
        else:
            return f"""FROM {base_image}

WORKDIR /app

COPY . .

EXPOSE {port}

CMD ["./run.sh"]
"""
    
    def _generate_docker_compose(self, config: Dict[str, Any]) -> str:
        """Generate docker-compose.yml content."""
        project_name = config.get("project_name", "app")
        port = config.get("docker", {}).get("port", 8080)
        
        return f"""version: '3.8'

services:
  {project_name}:
    build: .
    ports:
      - "{port}:{port}"
    environment:
      - NODE_ENV=production
    volumes:
      - .:/app
      - /app/node_modules
"""
    
    def _generate_dockerignore(self, language: str) -> str:
        """Generate .dockerignore content."""
        common_ignores = [
            ".git",
            ".gitignore",
            "README.md",
            "Dockerfile",
            "docker-compose.yml",
            ".dockerignore"
        ]
        
        language_ignores = {
            "python": ["__pycache__", "*.pyc", ".pytest_cache", ".venv"],
            "node": ["node_modules", "npm-debug.log"],
            "typescript": ["node_modules", "npm-debug.log", "dist"],
            "java": ["target", "*.class"],
            "kotlin": ["build", "*.class"]
        }
        
        ignores = common_ignores + language_ignores.get(language, [])
        return "\n".join(ignores) + "\n"


class DocsPlugin(ProjectPlugin):
    """Plugin for documentation setup."""
    
    @property
    def name(self) -> str:
        return "docs"
    
    @property
    def version(self) -> str:
        return "1.0.0"
    
    @property
    def description(self) -> str:
        return "Adds MkDocs documentation setup"
    
    def configure(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Configure docs plugin."""
        return config
    
    def generate_files(self, project_path: Path, config: Dict[str, Any]) -> None:
        """Generate documentation files."""
        if not config.get("features", {}).get("docs", False):
            return
        
        project_name = config.get("project_name", "Project")
        
        # Create docs directory
        docs_dir = project_path / "docs"
        docs_dir.mkdir(exist_ok=True)
        
        # Generate mkdocs.yml
        mkdocs_content = f"""site_name: {project_name}
nav:
  - Home: index.md
  - API: api.md
  - Contributing: contributing.md

theme:
  name: material
  features:
    - navigation.instant

markdown_extensions:
  - admonition
  - codehilite
"""
        (project_path / "mkdocs.yml").write_text(mkdocs_content)
        
        # Generate index.md
        index_content = f"""# {project_name}

Welcome to the {project_name} documentation.

## Getting Started

Add your getting started guide here.

## Features

- Feature 1
- Feature 2
- Feature 3
"""
        (docs_dir / "index.md").write_text(index_content)
        
        # Generate api.md
        api_content = """# API Reference

Document your API here.
"""
        (docs_dir / "api.md").write_text(api_content)
        
        # Generate contributing.md
        contributing_content = """# Contributing

Guidelines for contributing to this project.

## Development Setup

1. Clone the repository
2. Install dependencies
3. Run tests

## Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request
"""
        (docs_dir / "contributing.md").write_text(contributing_content)


class PluginManager:
    """Manages projgen plugins."""
    
    def __init__(self):
        self.plugins: Dict[str, ProjectPlugin] = {}
        self._load_builtin_plugins()
    
    def _load_builtin_plugins(self):
        """Load built-in plugins."""
        self.register_plugin(DockerPlugin())
        self.register_plugin(DocsPlugin())
    
    def register_plugin(self, plugin: ProjectPlugin):
        """Register a plugin."""
        self.plugins[plugin.name] = plugin
    
    def load_plugin_from_file(self, plugin_path: str) -> Optional[ProjectPlugin]:
        """
        Load a plugin from a Python file.
        
        Args:
            plugin_path: Path to the plugin file
            
        Returns:
            Loaded plugin instance or None if failed
        """
        try:
            spec = importlib.util.spec_from_file_location("plugin", plugin_path)
            module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module)
            
            # Look for plugin class
            for attr_name in dir(module):
                attr = getattr(module, attr_name)
                if (isinstance(attr, type) and 
                    issubclass(attr, ProjectPlugin) and 
                    attr != ProjectPlugin):
                    plugin = attr()
                    self.register_plugin(plugin)
                    return plugin
            
        except Exception as e:
            print(f"Failed to load plugin from {plugin_path}: {e}")
        
        return None
    
    def load_plugins_from_directory(self, plugins_dir: str):
        """Load all plugins from a directory."""
        plugins_path = Path(plugins_dir)
        if not plugins_path.exists():
            return
        
        for plugin_file in plugins_path.glob("*.py"):
            self.load_plugin_from_file(str(plugin_file))
    
    def get_plugin(self, name: str) -> Optional[ProjectPlugin]:
        """Get a plugin by name."""
        return self.plugins.get(name)
    
    def list_plugins(self) -> List[ProjectPlugin]:
        """List all registered plugins."""
        return list(self.plugins.values())
    
    def apply_plugins(self, project_path: Path, config: Dict[str, Any]):
        """Apply all relevant plugins to a project."""
        for plugin in self.plugins.values():
            try:
                # Configure plugin
                config = plugin.configure(config)
                
                # Generate files
                plugin.generate_files(project_path, config)
                
                # Post-generation hook
                plugin.post_generate(project_path, config)
                
            except Exception as e:
                print(f"Error applying plugin {plugin.name}: {e}")
    
    def get_plugin_info(self) -> str:
        """Get formatted information about all plugins."""
        if not self.plugins:
            return "No plugins available."
        
        info = ["Available Plugins:"]
        for plugin in self.plugins.values():
            info.append(f"  • {plugin.name} v{plugin.version}: {plugin.description}")
        
        return "\n".join(info)