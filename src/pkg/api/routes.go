package api

import (
	"runtime"

	fiber "github.com/gofiber/fiber/v2"
)

func (a *Api) Ping(c *fiber.Ctx) error {
	return c.SendString("pong")
}

func (a *Api) NodeInfo(c *fiber.Ctx) error {
	// txs := a.core.GetPoolTxs()
	info, err := a.core.GetNodeInfo()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(info)
}

// return status about the system. G count and memory
func (a *Api) Stats(c *fiber.Ctx) error {
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	gCnt := runtime.NumGoroutine()
	alloc := mem.Alloc / 1024 / 1024
	ret := map[string]uint64{
		"goroutines":   uint64(gCnt),
		"mem_alloc_mb": alloc,
	}
	return c.JSON(ret)
}
