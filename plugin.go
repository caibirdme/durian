package durian

import (
	// plug in server
	_ "github.com/caibirdme/durian/server"

	// plug in directives
	_ "github.com/caibirdme/durian/gzip"
	_ "github.com/caibirdme/durian/header"
	_ "github.com/caibirdme/durian/log"
	_ "github.com/caibirdme/durian/not_found"
	_ "github.com/caibirdme/durian/response"
	_ "github.com/caibirdme/durian/reverse_proxy"
	_ "github.com/caibirdme/durian/rewrite"
	_ "github.com/caibirdme/durian/root"
	_ "github.com/caibirdme/durian/status"
	_ "github.com/caibirdme/durian/timeout"
)
