module github.com/spiral/roadrunner/v2

go 1.15

require (
	github.com/buger/goterm v0.0.0-20200322175922-2f3e71b85129
	github.com/dustin/go-humanize v0.0.0-20171111073723-bb3d318650d4
	github.com/fatih/color v1.10.0
	github.com/json-iterator/go v1.1.10
	github.com/mattn/go-runewidth v0.0.9
	github.com/olekukonko/tablewriter v0.0.4
	github.com/shirou/gopsutil v3.20.11+incompatible
	github.com/spf13/cobra v1.1.1
	github.com/spiral/endure v1.0.0-beta20
	github.com/spiral/errors v1.0.6
	github.com/spiral/goridge/v3 v3.0.0-beta8
	github.com/spiral/roadrunner-plugins/config v1.0.0
	github.com/spiral/roadrunner-plugins/http v1.0.2
	github.com/spiral/roadrunner-plugins/informer v1.0.4
	github.com/spiral/roadrunner-plugins/logger v1.0.1
	github.com/spiral/roadrunner-plugins/metrics v1.0.0
	github.com/spiral/roadrunner-plugins/redis v1.0.0
	github.com/spiral/roadrunner-plugins/reload v1.0.1
	github.com/spiral/roadrunner-plugins/resetter v1.0.0
	github.com/spiral/roadrunner-plugins/rpc v1.0.1
	github.com/spiral/roadrunner-plugins/server v1.0.4
	github.com/stretchr/testify v1.6.1
	github.com/valyala/tcplisten v0.0.0-20161114210144-ceec8f93295a
	github.com/vbauerster/mpb/v5 v5.4.0
	go.uber.org/multierr v1.6.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
)

//replace github.com/spiral/roadrunner-plugins/http v1.0.2 => ../roadrunner-plugins/http
//replace github.com/spiral/roadrunner-plugins/reload v1.0.0 => ../roadrunner-plugins/reload
//replace github.com/spiral/roadrunner-plugins/rpc v1.0.0 => ../roadrunner-plugins/rpc
//replace github.com/spiral/roadrunner-plugins/server v1.0.3 => ../roadrunner-plugins/server
