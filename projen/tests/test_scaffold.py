import os
import shutil
import tempfile
import unittest

from config import DEFAULTS
from generators.scaffold import scaffold_project


class TestScaffoldProject(unittest.TestCase):
    def setUp(self):
        # Create a temporary directory for templates and projects
        self.temp_root = tempfile.mkdtemp()
        self.templates_dir = os.path.join(self.temp_root, "templates")
        # Create template subdirectories
        for sub in ("common", "gradle", "bazel", "ci"):
            os.makedirs(os.path.join(self.templates_dir, sub))

        # Write minimal dummy templates
        common = {
            "common/README.md.j2": "{{ project_name }}\n",
            "common/LICENSE.j2": "{{ license }}\n",
            "common/gitignore.j2": "*.pyc\n",
        }
        gradle = {
            "gradle/build.gradle.j2": "// Gradle build for {{ project_name }}\n",
            "gradle/settings.gradle.j2": "rootProject.name = '{{ project_name }}'\n",
        }
        bazel = {
            "bazel/WORKSPACE.j2": "# Workspace for {{ project_name }}\n",
            "bazel/BUILD.bazel.j2": "# Build for {{ project_name }}\n",
        }
        ci = {
            "ci/github.yml.j2": "# CI for {{ project_name }}\n",
        }
        for mapping in (common, gradle, bazel, ci):
            for relpath, content in mapping.items():
                path = os.path.join(self.templates_dir, relpath)
                with open(path, "w") as f:
                    f.write(content)

        # Override the templates directory in DEFAULTS
        self.original_templates_dir = DEFAULTS["templates_dir"]
        DEFAULTS["templates_dir"] = self.templates_dir

    def tearDown(self):
        # Restore DEFAULTS and clean up
        DEFAULTS["templates_dir"] = self.original_templates_dir
        shutil.rmtree(self.temp_root)

    def test_scaffold_both_builds(self):
        project_name = os.path.join(self.temp_root, "testproj")
        scaffold_project(
            project_name=project_name,
            build_system="both",
            language="python",
            license_id="Apache-2.0",
            ci_provider="github",
            overwrite=False,
        )
        # Common files
        self.assertTrue(os.path.isfile(os.path.join(project_name, "README.md")))
        self.assertTrue(os.path.isfile(os.path.join(project_name, "LICENSE")))
        self.assertTrue(os.path.isfile(os.path.join(project_name, ".gitignore")))
        # Gradle files
        self.assertTrue(os.path.isfile(os.path.join(project_name, "build.gradle")))
        self.assertTrue(os.path.isfile(os.path.join(project_name, "settings.gradle")))
        # Bazel files
        self.assertTrue(os.path.isfile(os.path.join(project_name, "WORKSPACE")))
        self.assertTrue(os.path.isfile(os.path.join(project_name, "BUILD.bazel")))
        # Source directories
        self.assertTrue(os.path.isdir(os.path.join(project_name, "src/main/python")))
        self.assertTrue(os.path.isdir(os.path.join(project_name, "src/test/python")))
        # CI workflow
        self.assertTrue(
            os.path.isfile(os.path.join(project_name, ".github", "workflows", "ci.yml"))
        )

    def test_overwrite_option(self):
        # Create existing directory
        project_dir = os.path.join(self.temp_root, "existing")
        os.makedirs(project_dir)
        # Without overwrite, should raise
        with self.assertRaises(FileExistsError):
            scaffold_project(
                project_name=project_dir,
                build_system="bazel",
                language="java",
                license_id="MIT",
                ci_provider="github",
                overwrite=False,
            )
        # With overwrite, should succeed
        scaffold_project(
            project_name=project_dir,
            build_system="bazel",
            language="java",
            license_id="MIT",
            ci_provider="github",
            overwrite=True,
        )
        # Bazel artifacts present
        self.assertTrue(os.path.isfile(os.path.join(project_dir, "WORKSPACE")))
        self.assertTrue(os.path.isfile(os.path.join(project_dir, "BUILD.bazel")))
        # No Gradle artifacts
        self.assertFalse(os.path.exists(os.path.join(project_dir, "build.gradle")))


if __name__ == "__main__":
    unittest.main()
