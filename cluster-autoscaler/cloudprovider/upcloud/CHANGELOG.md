# Changelog

All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## [Unreleased]

### Changed

- synced `feat/cluster-autoscaler-cloudprovider-upcloud` with upstream `master` branch (CA 1.36.0, k8s deps v1.36.1)
  - registered the provider via `init()` in `upcloud_cloud_provider.go` and removed the obsolete `builder_upcloud.go`, matching the new upstream registration model (`RegisterCloudProvider`/`SetDefaultCloudProvider`)
- bumped vendored UpCloud Go SDK to `v8.37.0`

### Fixed

- made node group scale-up non-blocking: `scaleNodeGroup` no longer polls the node group to `running` state, which blocked the main autoscaler loop for minutes on every scale-up; the target size is updated immediately and CA tracks upcoming nodes itself
- added retry with exponential backoff around UpCloud API calls (node group list/details and modify), as the UpCloud Go SDK has no built-in retry; retries cover request timeouts and `429`/`5xx` responses
- made node group cache refresh atomic: a transient API timeout no longer wipes the cache (which previously left the node group `unhealthy` with lost readiness tracking)
- added RBAC for `volumeattachments` (`storage.k8s.io`) and `resourceclaims`/`resourceslices`/`deviceclasses` (`resource.k8s.io`, DRA) required by the newer scheduler framework

## [1.2.0]

### Added

- added support for using the autoscaler with an UpCloud API token, via `UPCLOUD_TOKEN` env var

### Fixed

- synced `feat/cluster-autoscaler-cloudprovider-upcloud` with upstream `master` branch (CA 1.35.0-rc.0)
  - fixed regression with `*coreoptions.AutoscalerOptions` in UpCloud provider builder
  - fixed regression in `upCloudNodeGroup`: added `ForceDeleteNodes()` and delegating it to the existing `DeleteNodes()` flow
  - fixed regression in `upCloudNodeGroup`: updated `TemplateNodeInfo()` to match the current upstream `cloudprovider.NodeGroup` signature
- pinned the version of UpCloud autoscaler Docker image version in examples

### Images

| Kubernetes Version | Image                                 |
| ------------------ | ------------------------------------- |
| 1.29.x             | ghcr.io/upcloudltd/autoscaler:v1.29.5 |

## [1.1.0]

### Added

- set custom limits for node groups using `--nodes` parameter

### Changed

- synced `feat/cluster-autoscaler-cloudprovider-upcloud` with `master` branch (CA 1.31.0-beta.0)

### Images

| Kubernetes Version | Image                                 |
| ------------------ | ------------------------------------- |
| 1.29.x             | ghcr.io/upcloudltd/autoscaler:v1.29.4 |
| 1.28.x             | ghcr.io/upcloudltd/autoscaler:v1.28.6 |
| 1.27.X             | ghcr.io/upcloudltd/autoscaler:v1.27.8 |

## [1.0.0]

First stable release

[Unreleased]: https://github.com/UpCloudLtd/autoscaler/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/UpCloudLtd/autoscaler/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/UpCloudLtd/autoscaler/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/UpCloudLtd/autoscaler/releases/tag/v1.0.0
