"""
# Custom Bazel rules for building multi-language binaries
"""

def _multi_lang_binary_impl(ctx):
    output = ctx.actions.declare_file(ctx.label.name + ".zip")
    inputs = [f.path for f in ctx.files.srcs]

    # zip all binaries into one artifact
    ctx.actions.run_shell(
        outputs = [output],
        command = "zip -j {output} {input}".format(
            output = output.path,
            input = " ".join([f.path for f in ctx.files.srcs]),
        ),
    )
    return DefaultInfo(files = depset([output]))

multi_lang_binary = rule(
    implementation = _multi_lang_binary_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True),
    },
)
