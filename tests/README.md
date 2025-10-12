# Test Suite Layout

The legacy test suite still lives under `pkg/cgocopy` so that it can reach
internal helpers and CGO fixtures. As we lift more helpers into exported
wrappers we can migrate individual files into `tests/unit`, `tests/integration`,
and `tests/benchmarks`. The folders are provisioned so future work can land
without another repo-wide restructure.
