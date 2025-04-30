from migrate import migrate_gradle_to_bazel

def test_migrate_gradle_to_bazel(tmp_path):
    gradle = tmp_path / "build.gradle.kts"
    gradle.write_text(
        """
        plugins { id("application") }
        dependencies {
            implementation("org.jetbrains.kotlin:kotlin-stdlib:1.5.0")
        }
        """)

    out = tmp_path / "BUILD.bazel"
    migrate_gradle_to_bazel(str(gradle), str(out))
    content = out.read_text()
    assert "kotlin-stdlib:1.5.0" in content
    assert "kt_jvm_library" in content