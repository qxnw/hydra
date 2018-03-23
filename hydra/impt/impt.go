// +build !oci

package impt

import (
	_ "github.com/qxnw/hydra/client/rpc"
	_ "github.com/qxnw/hydra/engines"
	_ "github.com/qxnw/hydra/registry/local"
	_ "github.com/qxnw/hydra/registry/zookeeper"
	_ "github.com/qxnw/hydra/servers/cron"
	_ "github.com/qxnw/hydra/servers/http"
	_ "github.com/qxnw/hydra/servers/mqc"
	_ "github.com/qxnw/hydra/servers/rpc"
	_ "github.com/qxnw/lib4go/cache/memcache"
	_ "github.com/qxnw/lib4go/cache/redis"
	_ "github.com/qxnw/lib4go/mq/redis"
	_ "github.com/qxnw/lib4go/mq/stomp"
	_ "github.com/qxnw/lib4go/mq/xmq"
	_ "github.com/qxnw/lib4go/queue"
	_ "github.com/qxnw/lib4go/queue/redis"
)
