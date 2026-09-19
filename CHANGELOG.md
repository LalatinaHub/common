# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-09-19

### Fixed
- **subconverter**: Fixed sing-box configuration compatibility with v1.14.1
  - Removed deprecated `"type": "dns"` outbound (removed since sing-box v1.13+)
  - Fixed DNS hijacking rule to use logical OR rule (`protocol: dns OR port: 53`)
  - Added `sniff` action as first route rule for proper protocol detection
  - Updated tests to validate sing-box 1.14.1 compliance

### Changed
- **subconverter**: Output now generates 8 outbounds instead of 9 (removed DNS outbound)
- **subconverter**: Route rules now follow sing-box 1.14.1 specification

## [0.1.0] - 2024-XX-XX

### Added
- Initial release with core functionality
- Database connection pooling for Turso LibSQL
- Domain models (User, Server, KeyValue, ProxyNode)
- Repository interfaces and implementations
- Bi-directional proxy URL parser/formatter
- Multi-format subconverter (Clash Meta, sing-box, SFA/BFR, Base64)
- IATA airport geolocation lookup
- Network probe runners
- UDP packet relay functionality
