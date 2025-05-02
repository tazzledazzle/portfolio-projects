import os

DEFAULTS = {
    # Default build system: bazel, gradle, or both
    "build": "both",
    # Supported languages
    "languages": [
        "java",
        "kotlin",
        "groovy",
        "cpp",
        "c",
        "python",
        "rust",
        "node",
        "typescript",
    ],
    # CI providers
    "ci_providers": ["github"],
    # Default license
    "license": "MIT",
    # Directory where Jinja2 templates reside
    "templates_dir": os.path.join(os.path.dirname(__file__), "templates"),
}
