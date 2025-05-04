from setuptools import setup, find_packages

setup(
    name="projgen",
    version="1",
    description="CLI tool to scaffold Bazel/Gradle projects with opinionated defaults.",
    author="Terence Schumacher",
    author_email="terenceschumacher@gmail.com",
    url="https://github.com/tazzledazzle/projgen",
    packages=find_packages(exclude=["tests*", "templates*"]),
    include_package_data=True,
    install_requires=[
        "click>=8.0",
        "jinja2>=3.0",
    ],
    entry_points={
        "console_scripts": [
            "projgen=cli",
        ],
    },
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
)