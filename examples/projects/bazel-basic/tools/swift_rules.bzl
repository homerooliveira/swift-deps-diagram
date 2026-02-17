def _noop_impl(ctx):
    return []

swift_library = rule(
    implementation = _noop_impl,
    attrs = {
        "deps": attr.label_list(),
    },
)

swift_binary = rule(
    implementation = _noop_impl,
    attrs = {
        "deps": attr.label_list(),
    },
)

swift_test = rule(
    implementation = _noop_impl,
    test = True,
    attrs = {
        "deps": attr.label_list(),
    },
)
