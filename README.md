# Grout DRA driver

A kubelet Dynamic Resource Allocation driver for KubeVirt vhost-user devices.
It publishes `grout.openperouter.io` resources, prepares each claim's socket
mount and `vhost-user-path` metadata, and connects Grout to QEMU's socket.

The driver calls the Go control API from the
[`grout-cni`](https://github.com/oshoval/grout-cni) repository. Its module is
pinned to the `poc` commit `7c4db4a`; it does not execute the CNI binary.
The POC catalog advertises eight synthetic slots and gateway addresses; see
[POC addressing](docs/poc-addressing.md). Replace the catalog for real devices.

## Build and deploy

```bash
make image IMAGE=quay.io/oshoval/groutdra:dev
kubectl apply -f deploy/groutdra.yaml
```

The runtime image contains only the static driver binary. Grout runs in the
OpenPERouter router DaemonSet. Enable the chart value
`openperouter.grout.vhostUser=true` to share the QEMU vhost socket directory.
The driver DaemonSet mounts that directory and Grout's host control socket.
Enable KubeVirt's `NetworkDevicesWithDRA` feature gate and its `vhostuser`
binding separately.

The existing cookbook3 records the earlier in-tree POC. Its router, KubeVirt,
and VMI steps still apply, but build this image here, apply `deploy/groutdra.yaml`,
and use the Helm value above instead of patching the router DaemonSet. Cluster
validation of the standalone layout is pending.
