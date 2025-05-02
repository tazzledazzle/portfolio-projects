import os
import shutil
from jinja2 import Environment, FileSystemLoader
from config import DEFAULTS


def scaffold_project(
    project_name, build_system, language, license_id, ci_provider, overwrite=False
):
    """Create a new project scaffold with given parameters."""
    root = os.path.abspath(project_name)
    # Handle existing directory
    if os.path.exists(root):
        if overwrite:
            shutil.rmtree(root)
        else:
            raise FileExistsError(f"Project directory {root} already exists.")
    os.makedirs(root)

    # Jinja2 environment setup
    templates_path = DEFAULTS["templates_dir"]
    env = Environment(
        loader=FileSystemLoader(templates_path), trim_blocks=True, lstrip_blocks=True
    )

    # Common context for templates
    ctx = {
        "project_name": project_name,
        "build_system": build_system,
        "language": language,
        "license": license_id,
        "ci_provider": ci_provider,
    }

    # Render common files: README, LICENSE, .gitignore
    common = ["README.md.j2", "LICENSE.j2", "gitignore.j2"]
    for tmpl in common:
        tpl = env.get_template(f"common/{tmpl}")
        content = tpl.render(**ctx)
        filename = tmpl.replace(".j2", "")
        if filename == "gitignore":
            filename = ".gitignore"
        with open(os.path.join(root, filename), "w") as f:
            f.write(content)

    # Gradle support
    if build_system in ("gradle", "both"):
        # Build files
        for tmpl in ["build.gradle.j2", "settings.gradle.j2"]:
            tpl = env.get_template(f"gradle/{tmpl}")
            content = tpl.render(**ctx)
            fname = tmpl.replace(".j2", "")
            with open(os.path.join(root, fname), "w") as f:
                f.write(content)
        # Source directories
        os.makedirs(os.path.join(root, f"src/main/{language}"), exist_ok=True)
        os.makedirs(os.path.join(root, f"src/test/{language}"), exist_ok=True)

    # Bazel support
    if build_system in ("bazel", "both"):
        for tmpl in ["WORKSPACE.j2", "BUILD.bazel.j2"]:
            tpl = env.get_template(f"bazel/{tmpl}")
            content = tpl.render(**ctx)
            fname = tmpl.replace(".j2", "")
            with open(os.path.join(root, fname), "w") as f:
                f.write(content)
        os.makedirs(os.path.join(root, "src"), exist_ok=True)
        # Initial Bazel BUILD file
        with open(os.path.join(root, "src/BUILD.bazel"), "w") as f:
            f.write("# TODO: Add Bazel targets here\n")

    # CI integration
    ci_tmpl = env.get_template(f"ci/{ci_provider}.yml.j2")
    workflows = os.path.join(root, ".github", "workflows")
    os.makedirs(workflows, exist_ok=True)
    with open(os.path.join(workflows, "ci.yml"), "w") as f:
        f.write(ci_tmpl.render(**ctx))

    # Feedback
    print(f"Scaffolded {project_name} [{build_system} + {language}]")
