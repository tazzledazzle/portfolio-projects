import os

# Scaffold directories and files again after state reset
base_dir = "/mnt/data/devstack-manager"

dirs = [
    "backend/app",
    "backend/app/services",
    "frontend/src/components",
    "frontend/public",
    "frontend/src/pages",
]

files = {
    "backend/app/main.py": """\

""",
    "backend/app/services/docker_control.py": """\

""",
    "frontend/src/main.jsx": """\
""",
    "frontend/src/pages/App.jsx": """\
""",
    "frontend/index.html": """\
""",
    "frontend/src/index.css": """\

""",
    "docker-compose.yml": """\
""",
    "backend/Dockerfile": """\
""",
    "frontend/Dockerfile": """\
""",
    ".gitignore": """\
"""
}

# Create directories and files
os.makedirs(base_dir, exist_ok=True)
for d in dirs:
    os.makedirs(os.path.join(base_dir, d), exist_ok=True)
for path, content in files.items():
    with open(os.path.join(base_dir, path), "w") as f:
        f.write(content)

base_dir