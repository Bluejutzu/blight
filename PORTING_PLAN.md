# Cross-OS Porting Plan (Incremental)

This branch is dedicated to incremental, reviewable cross-platform work.

## Goal
Port the app cleanly to additional desktop OS targets with **small, focused commits** instead of large bundled changes.

## Target order
1. Linux
2. macOS

## Commit strategy
Each commit should contain one logical porting unit, for example:
- Build/toolchain config for one OS
- Runtime path handling update
- Platform-specific feature shim
- Packaging/signing config
- CI workflow update for one OS target

Avoid combining unrelated platform changes in a single commit.

## Definition of done per OS
- App builds successfully for the target OS.
- App launches and basic UI flow works.
- Any OS-specific caveats are documented.
- CI has at least one verification path for the OS.
