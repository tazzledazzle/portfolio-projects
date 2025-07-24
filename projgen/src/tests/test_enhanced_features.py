"""Tests for enhanced projgen features."""

import os
import tempfile
import unittest
from pathlib import Path

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from projgen.validation import TemplateValidator, ValidationError
from projgen.plugins import PluginManager, DockerPlugin, DocsPlugin


class TestValidation(unittest.TestCase):
    """Test validation functionality."""
    
    def setUp(self):
        self.temp_dir = tempfile.mkdtemp()
        self.validator = TemplateValidator(self.temp_dir)
    
    def test_validate_project_name(self):
        """Test project name validation."""
        # Valid names
        self.assertTrue(self.validator._validate_project_name("my-project"))
        self.assertTrue(self.validator._validate_project_name("MyProject123"))
        self.assertTrue(self.validator._validate_project_name("project_name"))
        
        # Invalid names
        self.assertFalse(self.validator._validate_project_name(""))
        self.assertFalse(self.validator._validate_project_name("123project"))
        self.assertFalse(self.validator._validate_project_name("my project"))
    
    def test_validate_language(self):
        """Test language validation."""
        self.assertTrue(self.validator._validate_language("python"))
        self.assertTrue(self.validator._validate_language("java"))
        self.assertFalse(self.validator._validate_language("invalid"))
        self.assertFalse(self.validator._validate_language(""))
    
    def test_validate_build_system(self):
        """Test build system validation."""
        self.assertTrue(self.validator._validate_build_system("gradle"))
        self.assertTrue(self.validator._validate_build_system("bazel"))
        self.assertTrue(self.validator._validate_build_system("both"))
        self.assertFalse(self.validator._validate_build_system("maven"))
        self.assertFalse(self.validator._validate_build_system(""))
    
    def test_validate_project_config(self):
        """Test full project configuration validation."""
        valid_config = {
            "project_name": "test-project",
            "language": "python",
            "build_system": "gradle",
            "license_id": "MIT"
        }
        self.assertTrue(self.validator.validate_project_config(valid_config))
        
        invalid_config = {
            "project_name": "123invalid",
            "language": "invalid",
            "build_system": "invalid",
            "license_id": "MIT"
        }
        self.assertFalse(self.validator.validate_project_config(invalid_config))


class TestPlugins(unittest.TestCase):
    """Test plugin functionality."""
    
    def setUp(self):
        self.temp_dir = tempfile.mkdtemp()
        self.project_path = Path(self.temp_dir) / "test-project"
        self.project_path.mkdir()
        self.plugin_manager = PluginManager()
    
    def test_docker_plugin(self):
        """Test Docker plugin functionality."""
        docker_plugin = DockerPlugin()
        
        config = {
            "project_name": "test-project",
            "language": "python",
            "features": {"docker": True}
        }
        
        # Configure plugin
        updated_config = docker_plugin.configure(config)
        self.assertIn("docker", updated_config)
        self.assertEqual(updated_config["docker"]["base_image"], "python:3.12-slim")
        
        # Generate files
        docker_plugin.generate_files(self.project_path, updated_config)
        
        # Check generated files
        self.assertTrue((self.project_path / "Dockerfile").exists())
        self.assertTrue((self.project_path / "docker-compose.yml").exists())
        self.assertTrue((self.project_path / ".dockerignore").exists())
    
    def test_docs_plugin(self):
        """Test documentation plugin functionality."""
        docs_plugin = DocsPlugin()
        
        config = {
            "project_name": "test-project",
            "features": {"docs": True}
        }
        
        # Generate files
        docs_plugin.generate_files(self.project_path, config)
        
        # Check generated files
        self.assertTrue((self.project_path / "mkdocs.yml").exists())
        self.assertTrue((self.project_path / "docs" / "index.md").exists())
        self.assertTrue((self.project_path / "docs" / "api.md").exists())
        self.assertTrue((self.project_path / "docs" / "contributing.md").exists())
    
    def test_plugin_manager(self):
        """Test plugin manager functionality."""
        # Check built-in plugins are loaded
        self.assertIsNotNone(self.plugin_manager.get_plugin("docker"))
        self.assertIsNotNone(self.plugin_manager.get_plugin("docs"))
        
        # List plugins
        plugins = self.plugin_manager.list_plugins()
        self.assertEqual(len(plugins), 2)
        
        # Get plugin info
        info = self.plugin_manager.get_plugin_info()
        self.assertIn("docker", info)
        self.assertIn("docs", info)


if __name__ == "__main__":
    unittest.main()