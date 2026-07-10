package handlers

import (
	"net"
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/store"
	"github.com/bestruirui/bestsub/internal/utils"
	"github.com/bestruirui/bestsub/pkg/mihomo"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/setting").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(settingList),
		).
		AddRoute(
			router.NewRoute("/update", http.MethodPut).
				Handle(settingSet),
		).
		AddRoute(
			router.NewRoute("/interfaces", http.MethodGet).
				Handle(settingInterfaces),
		)
}

// settingInterfaces 返回本机网卡列表
func settingInterfaces(c *gin.Context) {
	ifaces, err := net.Interfaces()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrInternalServer)
		return
	}
	type ifaceInfo struct {
		Name string `json:"name"`
		IP   string `json:"ip"`
	}
	result := make([]ifaceInfo, 0, len(ifaces))
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}
		// 取第一个非 IPv6 地址作为主 IP；若全是 IPv6 则取第一个
		ip := ""
		for _, addr := range addrs {
			ipStr := addr.(*net.IPNet).IP.String()
			if addr.(*net.IPNet).IP.To4() != nil {
				ip = ipStr
				break
			}
			if ip == "" {
				ip = ipStr
			}
		}
		if ip == "" {
			continue
		}
		result = append(result, ifaceInfo{Name: iface.Name, IP: ip})
	}
	resp.Success(c, result)
}

// settingList 返回所有配置项
func settingList(c *gin.Context) {
	resp.Success(c, store.SettingList())
}

// settingSet 接收 { "key": "...", "value": "..." } 更新单个配置项
func settingSet(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.SettingSet(req.Key, req.Value); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}

	// DNS 配置变更时立即生效
	if req.Key == model.SettingDNSDefault || req.Key == model.SettingDNSMain {
		mihomo.UpdateDNSConfig(utils.SplitComma(store.SettingGet(model.SettingDNSDefault)), utils.SplitComma(store.SettingGet(model.SettingDNSMain)))
	}

	resp.Success(c, nil)
}
