package api

import (
	"os"
	"runtime"

	fiber "github.com/gofiber/fiber/v2"
)

func (a *Api) Ping(c *fiber.Ctx) error {
	return c.SendString("pong")
}

type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
}

func (a *Api) Version(c *fiber.Ctx) error {
	ret := VersionResponse{
		Version:   os.Getenv("BUILD_VERSION"),
		BuildTime: os.Getenv("BUILD_TIME"),
	}
	return c.JSON(ret)
}

func (a *Api) NodeInfo(c *fiber.Ctx) error {
	// txs := a.core.GetPoolTxs()
	info, err := a.core.GetNodeInfo()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(info)
}

type StatsResponse struct {
	Goroutines int    `json:"goroutines"`
	MemAllocMb uint64 `json:"mem_alloc_mb"`
}

// @Summary Some status about the system. G count and memory
// @Description Get information about the current state of the system memory
// @Tags etc
// @Accept  json
// @Produce  json
// @Success 200 {object} StatsResponse
// @Failure 500 {object} APIError
// @Router /stats [get]
func (a *Api) Stats(c *fiber.Ctx) error {
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	gCnt := runtime.NumGoroutine()
	alloc := mem.Alloc / 1024 / 1024
	ret := StatsResponse{
		Goroutines: gCnt,
		MemAllocMb: alloc,
	}
	return apiSuccess(c, ret)
}
