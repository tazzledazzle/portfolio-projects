from __future__ import annotations


def render_pipeline(service_name: str, deploy_env: str) -> str:
    return f"""name: {service_name}-pipeline
on: [push]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - run: echo lint
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo test
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo build
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: echo deploy {deploy_env}
"""


def main() -> None:
    print(render_pipeline("checkout-service", "staging"))


if __name__ == "__main__":
    main()
