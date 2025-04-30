import re

def migrate_gradle_to_bazel(input_path: str, output_path: str):
    """
    Very basic migration function to convert a Gradle build file to a Bazel build file.
    reads build.gradle.kts for implementation deps and emits a Bazel kt_jvm_library stub

    Args:
        input_path (str): Path to the Gradle build file.
        output_path (str): Path to the Bazel build file.
    """
    deps = []
    with open(input_path, 'r') as f:
        for line in f:
            m = re.match(r"\s*implementation\("([\w\.-]+:[\w\.-]+:[\d\.]+)"\)", line)
            if m:
                deps.append(m.group(1))
    with open(output_path, 'w') as out:
        out.write('load("@io_bazel_rules_kotlin//kotlin:kotlin.bzl", "kt_jvm_library")\n')
        out.write('kt_jvm_library(\n')
        out.write('    name = "kotlin_lib",\n')
        out.write('    srcs = glob([\"src/main/kotlin/**/*.kt\"]),\n')
        if deps:
            out.write('    deps = [\n')
            for dep in deps:
                out.write(f'        "{dep}",\n')
            out.write('    ],\n')
        out.write(')\n')

def main():
    parser = argparse.ArgumentParser(description="Migrate Gradle Kotlin DSL build file to Bazel build file.")
    parser.add_argument("input", help="Path to the input Gradle build file.")
    parser.add_argument("output", help="Path to the output Bazel build file.")
    args = parser.parse_args()
    migrate_gradle_to_bazel(args.input, args.output)


if __name__ == "__main__":
    main()