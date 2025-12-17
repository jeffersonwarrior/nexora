## v0.28.0 (2025-12-17) - Maintenance Release

**Build & Stability Fixes:**
- Fix catwalk Provider struct compatibility (`ExtraBody` removal)
- Resolve logo rendering issues (imports, undefined symbols, emoji parsing)  
- Clean up undefined providers & references
- Fix AgentCoordinator.Run() signature in chat.go
- Remove RentalH200 provider (under rework) - all references cleaned

**Status:** Clean build, `./install.sh` verified, v0.28.0 tagged ✅