package ssh

import (
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh/agent"
)

// agentConn 缓存 agent 连接，进程内只建一次。
var (
	agentOnce   sync.Once
	agentClient agent.Agent
	agentConn   net.Conn
)

// getAgent 返回 SSH Agent 客户端。
//
// 通过 SSH_AUTH_SOCK 环境变量连接 agent。
// 未设置或连接失败时返回 nil。
func getAgent() agent.Agent {
	agentOnce.Do(func() {
		sock := os.Getenv("SSH_AUTH_SOCK")
		if sock == "" {
			return
		}

		conn, err := net.Dial("unix", sock)
		if err != nil {
			return
		}

		agentConn = conn
		agentClient = agent.NewClient(conn)
	})

	return agentClient
}
