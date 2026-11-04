# Builds the app the way the release workflow does.
#
# The with_gvisor tag puts the gVisor network stack in the core: without it
# the core has no stack for the TUN mode, and says so when the TUN is
# switched on.
$env:GOFLAGS = "-tags=with_gvisor"
go tool mygo build -platform windows/amd64 -skip-installer
