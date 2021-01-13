# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]
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

[Unreleased]: https://github.com/thousandeyes/shoelaces/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/thousandeyes/shoelaces/compare/v1.0.2...v1.1.0
[1.0.2]: https://github.com/thousandeyes/shoelaces/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/thousandeyes/shoelaces/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/thousandeyes/shoelaces/tree/v1.0.0
