---
trigger: always_on
---

# Common Pitfalls

1. **Don't assume DNF**: Use `util.PackageBinary()` - older systems have YUM
2. **Config file path**: Default is `/etc/txlog.yaml`, not in repo root
3. **API key compatibility**: Server < 1.14.0 will reject API keys - check in `initConfig()`
4. **RPM package building**: Requires `nfpm` and `pandoc` installed (see README dev setup)
5. **Static binary requirement**: Production needs CGO disabled for portability across RHEL versions
