"Custom build setting to parse a string flag"

BuildSettingInfo = provider(doc = "", fields = ["value"])

def _string_imp(ctx):
    value = ctx.build_setting_value
    label = ctx.label.name

    #buildifier: disable=print
    print("Evaluated value for " + label + ": " + value)
    return BuildSettingInfo(value = value)

string_flag = rule(
    implementation = _string_imp,
    # https://bazel.build/extending/config#the-build_setting-rule-parameter
    build_setting = config.string(flag = True),
)
