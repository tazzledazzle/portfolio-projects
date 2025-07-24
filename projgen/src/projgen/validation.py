"""Template validation for projgen."""

import re
from pathlib import Path
from typing import List, Dict, Any
from jinja2 import Environment, FileSystemLoader, TemplateSyntaxError
import yaml


class ValidationError(Exception):
    """Custom exception for validation errors."""
    pass


class TemplateValidator:
    """Validates project templates and configurations."""
    
    def __init__(self, templates_dir: str):
        self.templates_dir = Path(templates_dir)
        self.errors: List[str] = []
        self.warnings: List[str] = []
    
    def validate_project_config(self, config: Dict[str, Any]) -> bool:
        """
        Validate project configuration.
        
        Args:
            config: Project configuration dictionary
            
        Returns:
            True if valid, False otherwise
        """
        self.errors.clear()
        self.warnings.clear()
        
        # Validate project name
        if not self._validate_project_name(config.get("project_name", "")):
            return False
        
        # Validate language
        if not self._validate_language(config.get("language", "")):
            return False
        
        # Validate build system
        if not self._validate_build_system(config.get("build_system", "")):
            return False
        
        # Validate license
        if not self._validate_license(config.get("license_id", "")):
            return False
        
        return len(self.errors) == 0
    
    def _validate_project_name(self, name: str) -> bool:
        """Validate project name."""
        if not name:
            self.errors.append("Project name is required")
            return False
        
        if not re.match(r'^[a-zA-Z][a-zA-Z0-9_-]*$', name):
            self.errors.append(
                "Project name must start with a letter and contain only "
                "letters, numbers, hyphens, and underscores"
            )
            return False
        
        if len(name) > 50:
            self.warnings.append("Project name is quite long (>50 chars)")
        
        return True
    
    def _validate_language(self, language: str) -> bool:
        """Validate programming language."""
        from .config import DEFAULTS
        
        if not language:
            self.errors.append("Programming language is required")
            return False
        
        if language not in DEFAULTS["languages"]:
            self.errors.append(f"Unsupported language: {language}")
            return False
        
        return True
    
    def _validate_build_system(self, build_system: str) -> bool:
        """Validate build system."""
        valid_systems = ["bazel", "gradle", "both"]
        
        if not build_system:
            self.errors.append("Build system is required")
            return False
        
        if build_system not in valid_systems:
            self.errors.append(f"Invalid build system: {build_system}")
            return False
        
        return True
    
    def _validate_license(self, license_id: str) -> bool:
        """Validate license identifier."""
        if not license_id:
            self.warnings.append("No license specified")
            return True
        
        # Common license patterns
        common_licenses = [
            "MIT", "Apache-2.0", "GPL-3.0", "BSD-3-Clause", 
            "Unlicense", "ISC", "LGPL-2.1", "MPL-2.0"
        ]
        
        if license_id not in common_licenses:
            self.warnings.append(f"Uncommon license: {license_id}")
        
        return True
    
    def validate_templates(self) -> bool:
        """
        Validate all Jinja2 templates in the templates directory.
        
        Returns:
            True if all templates are valid, False otherwise
        """
        self.errors.clear()
        self.warnings.clear()
        
        if not self.templates_dir.exists():
            self.errors.append(f"Templates directory not found: {self.templates_dir}")
            return False
        
        # Find all template files
        template_files = list(self.templates_dir.rglob("*.j2"))
        
        if not template_files:
            self.warnings.append("No template files found")
            return True
        
        # Validate each template
        env = Environment(loader=FileSystemLoader(str(self.templates_dir)))
        
        for template_file in template_files:
            try:
                relative_path = template_file.relative_to(self.templates_dir)
                template = env.get_template(str(relative_path))
                
                # Try to render with dummy variables
                dummy_vars = self._get_dummy_template_vars()
                template.render(**dummy_vars)
                
            except TemplateSyntaxError as e:
                self.errors.append(f"Template syntax error in {relative_path}: {e}")
            except Exception as e:
                self.warnings.append(f"Template warning in {relative_path}: {e}")
        
        return len(self.errors) == 0
    
    def _get_dummy_template_vars(self) -> Dict[str, Any]:
        """Get dummy variables for template validation."""
        return {
            "project_name": "test_project",
            "language": "python",
            "license": "MIT",
            "build_system": "gradle",
            "ci_provider": "github",
            "author": "Test Author",
            "email": "test@example.com",
            "year": "2025"
        }
    
    def validate_template_manifest(self, manifest_path: str) -> bool:
        """
        Validate a template manifest file.
        
        Args:
            manifest_path: Path to the manifest YAML file
            
        Returns:
            True if valid, False otherwise
        """
        try:
            with open(manifest_path, 'r') as f:
                manifest = yaml.safe_load(f)
            
            # Required fields
            required_fields = ["name", "description", "version", "templates"]
            for field in required_fields:
                if field not in manifest:
                    self.errors.append(f"Missing required field in manifest: {field}")
            
            # Validate templates list
            if "templates" in manifest:
                for template in manifest["templates"]:
                    if not isinstance(template, dict):
                        self.errors.append("Template entries must be objects")
                        continue
                    
                    if "source" not in template or "target" not in template:
                        self.errors.append("Template entries must have 'source' and 'target'")
            
            return len(self.errors) == 0
            
        except yaml.YAMLError as e:
            self.errors.append(f"Invalid YAML in manifest: {e}")
            return False
        except FileNotFoundError:
            self.errors.append(f"Manifest file not found: {manifest_path}")
            return False
    
    def get_validation_report(self) -> str:
        """Get a formatted validation report."""
        report = []
        
        if self.errors:
            report.append("❌ Errors:")
            for error in self.errors:
                report.append(f"  • {error}")
        
        if self.warnings:
            report.append("⚠️  Warnings:")
            for warning in self.warnings:
                report.append(f"  • {warning}")
        
        if not self.errors and not self.warnings:
            report.append("✅ All validations passed!")
        
        return "\n".join(report)