# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [1.4.0] - 2026-06-05
### Added
- CI workflow running unit and integration tests on PRs and pushes to master.
- Release workflow producing reproducible `linux/amd64` and `linux/arm64` binaries on version tags, with `sha256sums.txt`. Builds are Debian-compatible (CGO_ENABLED=0, -trimpath, SOURCE_DATE_EPOCH).
- Version string embedded at build time via `-ldflags "-X main.version=..."`, logged at startup.
- Example templates for Flatcar Linux (Ignition v3), AlmaLinux (kickstart), Ubuntu, and Debian.
- Local CSS replacing Bootstrap. The interface is visually equivalent; custom stylesheets referencing Bootstrap classes will need updating.
- `make run` target for quickly starting a local test instance.

### Changed
- Upgraded `gopkg.in/yaml.v2` to `yaml.v3`; dropped `vendor/`.
- Replaced deprecated `io/ioutil` with `os` equivalents throughout.
- Replaced third-party dependencies (gorilla/mux, alice, namsral/flag, go-kit/log) with stdlib equivalents (net/http, log/slog).
- Updated copyright year to 2018-2026.

### Removed
- Bootstrap CSS, Bootstrap JS, jQuery, and Glyphicon fonts.
- CoreOS and CentOS example templates.

## [1.3.2] - 2022-09-05
### Added
- Test for `/start` endpoint.
- Support for custom parameters in integration tests.

## [1.3.1] - 2022-01-01
### Changed
- Updated to Go 1.19 and tidied dependencies.

## [1.3.0] - 2021-09-01
### Added
- Human-friendly entry point.
- iPXE executable support.

### Fixed
- No extra poll in ipxemenu.
- Python 3 compatibility in integration tests.

## [1.2.0] - 2021-01-13
### Added
- Use Go Modules instead of Go Dep.
- Add docker build.
- Add parameter `-base-url`.

### Changed
- Changing parameters `-port` and `-domain` in favor of `-bind-addr`.

## [1.1.0] - 2020-04-01
### Added
- Build production binaries.
- Support config subdirectories.

## [1.0.2] - 2019-01-09
### Added
- Add scdoc manpage.

### Changed
- Rename example-templates-configs to data-dir.
- Move screenshots to docs/screenshots.

## [1.0.1] - 2018-12-04
### Added
- Added dnsmasq example.

### Fixed
- Fixes in ipxe for Debian and CentOS.

## [1.0.0] - 2018-08-03
### Added
- First release.

[Unreleased]: https://github.com/thousandeyes/shoelaces/compare/v1.4.0...HEAD
[1.4.0]: https://github.com/thousandeyes/shoelaces/compare/v1.3.2...v1.4.0
[1.3.2]: https://github.com/thousandeyes/shoelaces/compare/v1.3.1...v1.3.2
[1.3.1]: https://github.com/thousandeyes/shoelaces/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/thousandeyes/shoelaces/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/thousandeyes/shoelaces/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/thousandeyes/shoelaces/compare/v1.0.2...v1.1.0
[1.0.2]: https://github.com/thousandeyes/shoelaces/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/thousandeyes/shoelaces/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/thousandeyes/shoelaces/tree/v1.0.0
