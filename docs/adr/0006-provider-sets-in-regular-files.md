# Provider sets live in regular source files

Each provider package exports `ProviderSet` from its main source file without the `wireinject` build tag, so it shows up in godoc and editors without extra build flags. The only cost is linking `github.com/google/wire`, whose runtime package has no imports. Only an application's injector file keeps `//go:build wireinject`.
