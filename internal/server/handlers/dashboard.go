package handlers

import (
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/gin-gonic/gin"
)

type dashboardData struct {
	TotalNodes         int                  `json:"total_nodes"`         // 订阅节点池和单独节点的总数。
	SubscriptionsTotal int                  `json:"subscriptions_total"` // 订阅总数。
	TasksTotal         int                  `json:"tasks_total"`         // 任务总数。
	SharesTotal        int                  `json:"shares_total"`        // 分享总数。
	Subscriptions      []model.Subscription `json:"subscriptions"`       // 已补充实时节点数的订阅列表。
	CountryCounts      map[string]int       `json:"country_counts"`      // ISO 国家代码到节点数的映射。
}

func init() {
	router.NewGroupRouter("/api/v1/dashboard").
		Use(middleware.Auth()).
		AddRoute(router.NewRoute("/summary", http.MethodGet).Handle(dashboardSummary))
}

func dashboardSummary(c *gin.Context) {
	subscriptions := store.SubscriptionList()
	standaloneNodes := store.NodeList()
	data := dashboardData{
		TotalNodes:         store.NodeLen(),
		SubscriptionsTotal: store.SubscriptionLen(),
		TasksTotal:         store.TaskLen(),
		SharesTotal:        store.ShareLen(),
		Subscriptions:      subscriptions,
		CountryCounts:      map[string]int{},
	}

	addCountry := func(country string) {
		if country != "" {
			data.CountryCounts[country]++
		}
	}
	for _, current := range standaloneNodes {
		addCountry(current.CountryCode)
	}
	for i := range data.Subscriptions {
		subscription := &data.Subscriptions[i]
		currentNodes := node.PoolListBySubscription(subscription.ID)
		subscription.NodeNum = uint32(len(currentNodes))
		data.TotalNodes += len(currentNodes)
		for _, current := range currentNodes {
			addCountry(current.Info.CountryCode)
		}
	}
	resp.Success(c, data)
}
