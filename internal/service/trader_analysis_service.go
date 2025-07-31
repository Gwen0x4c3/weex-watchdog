package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"weex-watchdog/pkg/logger"
	"weex-watchdog/pkg/weex"
)

// RedisClient 简化的Redis客户端接口
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

// TraderAnalysisService 交易员分析服务
type TraderAnalysisService struct {
	redisClient RedisClient
	logger      *logger.Logger
}

// NewTraderAnalysisService 创建交易员分析服务
func NewTraderAnalysisService(redisClient RedisClient, logger *logger.Logger) *TraderAnalysisService {
	return &TraderAnalysisService{
		redisClient: redisClient,
		logger:      logger,
	}
}

// TraderAnalysisResult 交易员分析结果
type TraderAnalysisResult struct {
	TraderID    string    `json:"trader_id"`
	TraderName  string    `json:"trader_name"`
	AnalyzeTime time.Time `json:"analyze_time"`
	TimeRange   string    `json:"time_range"`

	// 基础统计
	TotalOrders int     `json:"total_orders"`
	WinOrders   int     `json:"win_orders"`
	LoseOrders  int     `json:"lose_orders"`
	WinRate     float64 `json:"win_rate"`

	// 盈亏统计
	TotalPnl        float64 `json:"total_pnl"`
	AvgPnl          float64 `json:"avg_pnl"`
	MaxProfit       float64 `json:"max_profit"`
	MaxLoss         float64 `json:"max_loss"`
	ProfitLossRatio float64 `json:"profit_loss_ratio"`

	// 风险指标
	AvgHoldingHours float64 `json:"avg_holding_hours"`
	MaxHoldingHours float64 `json:"max_holding_hours"`
	HoldingTimeP25  float64 `json:"holding_time_p25"`
	HoldingTimeP50  float64 `json:"holding_time_p50"` // 中位数
	HoldingTimeP75  float64 `json:"holding_time_p75"`

	// 币种分析
	CoinFrequency map[string]int     `json:"coin_frequency"`
	CoinWinRate   map[string]float64 `json:"coin_win_rate"`
	CoinPnl       map[string]float64 `json:"coin_pnl"`

	// 杠杆分析
	LeverageStats map[string]int `json:"leverage_stats"`

	// 时间分布
	HourlyStats map[int]int    `json:"hourly_stats"`
	DailyStats  map[string]int `json:"daily_stats"`

	// 持仓时间分布
	HoldingTimeDistribution []HoldingTimeBucket `json:"holding_time_distribution"`

	// 跟投收益模拟
	FollowProfit *FollowProfitResult `json:"follow_profit,omitempty"`
}

// HoldingTimeBucket 持仓时间分布桶
type HoldingTimeBucket struct {
	Range string  `json:"range"`
	Count int     `json:"count"`
	Rate  float64 `json:"rate"`
}

// FollowProfitResult 跟投收益结果
type FollowProfitResult struct {
	InitialCapital float64 `json:"initial_capital"`
	InvestPerOrder float64 `json:"invest_per_order"`
	FinalCapital   float64 `json:"final_capital"`
	TotalProfit    float64 `json:"total_profit"`
	ProfitRate     float64 `json:"profit_rate"`
	OrdersFollowed int     `json:"orders_followed"`
	MaxDrawdown    float64 `json:"max_drawdown"` // 最大回撤率

	CapitalCurve []CapitalDataPoint `json:"capital_curve"` // 资金曲线
}

// CapitalDataPoint 资金曲线数据点
type CapitalDataPoint struct {
	Time    time.Time `json:"time"`
	Capital float64   `json:"capital"`
}

// AnalyzeTrader 分析交易员历史数据
func (s *TraderAnalysisService) AnalyzeTrader(traderID string, timeRange string, initialCapital, investPerOrder float64) (*TraderAnalysisResult, error) {
	// 检查缓存
	cacheKey := fmt.Sprintf("trader_analysis:%s:%s", traderID, timeRange)
	cached, err := s.redisClient.Get(context.Background(), cacheKey)
	if err == nil {
		var result TraderAnalysisResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			// 如果有跟投参数，重新计算跟投收益
			if initialCapital > 0 && investPerOrder > 0 {
				result.FollowProfit = s.calculateFollowProfit(result.TraderID, timeRange, initialCapital, investPerOrder)
			}
			return &result, nil
		}
	}

	// 取历weex上历史订单数据
	orders, err := weex.GetHistoryOrderList(traderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history orders: %w", err)
	}

	// 根据时间范围过滤订单
	filteredOrders := s.filterOrdersByTimeRange(orders, timeRange)
	if len(filteredOrders) == 0 {
		return nil, fmt.Errorf("no orders found in the specified time range")
	}

	// 进行分析
	result := s.analyzeOrders(traderID, filteredOrders, timeRange)

	// 计算跟投收益
	if initialCapital > 0 && investPerOrder > 0 {
		result.FollowProfit = s.calculateFollowProfit(traderID, timeRange, initialCapital, investPerOrder)
	}

	// 缓存结果（30分钟）
	resultBytes, _ := json.Marshal(result)
	s.redisClient.Set(context.Background(), cacheKey, string(resultBytes), 30*time.Minute)

	return result, nil
}

// filterOrdersByTimeRange 根据时间范围过滤订单
func (s *TraderAnalysisService) filterOrdersByTimeRange(orders []weex.OpenOrder, timeRange string) []weex.OpenOrder {
	if timeRange == "" || timeRange == "all" {
		return orders
	}

	now := time.Now()
	var startTime time.Time

	switch timeRange {
	case "7d":
		startTime = now.AddDate(0, 0, -7)
	case "30d":
		startTime = now.AddDate(0, 0, -30)
	case "90d":
		startTime = now.AddDate(0, 0, -90)
	case "1y":
		startTime = now.AddDate(-1, 0, 0)
	default:
		return orders
	}

	var filtered []weex.OpenOrder
	for _, order := range orders {
		timestamp, err := strconv.ParseInt(order.OpenTime, 10, 64)
		if err != nil {
			continue
		}
		// 将毫秒时间戳转换为秒级时间戳
		orderTime := time.Unix(timestamp/1000, 0)
		if orderTime.After(startTime) {
			filtered = append(filtered, order)
		}
	}

	return filtered
}

// analyzeOrders 分析订单数据
func (s *TraderAnalysisService) analyzeOrders(traderID string, orders []weex.OpenOrder, timeRange string) *TraderAnalysisResult {
	result := &TraderAnalysisResult{
		TraderID:      traderID,
		AnalyzeTime:   time.Now(),
		TimeRange:     timeRange,
		CoinFrequency: make(map[string]int),
		CoinWinRate:   make(map[string]float64),
		CoinPnl:       make(map[string]float64),
		LeverageStats: make(map[string]int),
		HourlyStats:   make(map[int]int),
		DailyStats:    make(map[string]int),
	}

	if len(orders) == 0 {
		return result
	}

	result.TraderName = orders[0].TraderName
	result.TotalOrders = len(orders)

	var totalPnl float64
	var profits, losses, holdingTimes []float64
	coinWins := make(map[string]int)
	coinTotals := make(map[string]int)
	coinPnls := make(map[string]float64)
	contractMapper := weex.GetContractMapper()

	for _, order := range orders {
		// 解析盈亏
		pnl, _ := strconv.ParseFloat(order.RealizedPnl, 64)
		totalPnl += pnl

		// 统计胜负
		if pnl > 0 {
			result.WinOrders++
			profits = append(profits, pnl)
		} else if pnl < 0 {
			result.LoseOrders++
			losses = append(losses, math.Abs(pnl))
		}

		// 币种统计
		contractSymbol := contractMapper.GetSymbolName(order.ContractID)
		result.CoinFrequency[contractSymbol]++
		coinTotals[contractSymbol]++
		coinPnls[contractSymbol] += pnl
		if pnl > 0 {
			coinWins[contractSymbol]++
		}

		// 杠杆统计
		leverage := order.OpenLeverage
		result.LeverageStats[leverage]++

		// 持仓时间计算
		openTimestamp, err := strconv.ParseInt(order.OpenTime, 10, 64)
		if err != nil {
			continue
		}
		// 将毫秒时间戳转换为秒级时间戳
		openTime := time.Unix(openTimestamp/1000, 0)

		var holdingHours float64
		if order.CloseTime != "" {
			closeTimestamp, err := strconv.ParseInt(order.CloseTime, 10, 64)
			if err == nil {
				// 将毫秒时间戳转换为秒级时间戳
				closeTime := time.Unix(closeTimestamp/1000, 0)
				holdingHours = closeTime.Sub(openTime).Hours()
				holdingTimes = append(holdingTimes, holdingHours)
			}
		} // 时间分布统计
		result.HourlyStats[openTime.Hour()]++
		dateStr := openTime.Format("2006-01-02")
		result.DailyStats[dateStr]++
	}

	// 计算基础指标
	result.TotalPnl = totalPnl
	result.AvgPnl = totalPnl / float64(result.TotalOrders)
	if result.TotalOrders > 0 {
		result.WinRate = float64(result.WinOrders) / float64(result.TotalOrders) * 100
	}

	// 计算最大盈利和亏损
	if len(profits) > 0 {
		result.MaxProfit = s.max(profits)
	}
	if len(losses) > 0 {
		result.MaxLoss = s.max(losses)
	}

	// 计算盈亏比
	if len(profits) > 0 && len(losses) > 0 {
		avgProfit := s.average(profits)
		avgLoss := s.average(losses)
		if avgLoss > 0 {
			result.ProfitLossRatio = avgProfit / avgLoss
		}
	}

	// 计算风险指标
	if len(holdingTimes) > 0 {
		result.AvgHoldingHours = s.average(holdingTimes)
		result.MaxHoldingHours = s.max(holdingTimes)
		result.HoldingTimeP25 = s.percentile(holdingTimes, 25)
		result.HoldingTimeP50 = s.percentile(holdingTimes, 50)
		result.HoldingTimeP75 = s.percentile(holdingTimes, 75)
	}

	// 计算各币种胜率和盈亏
	for coin, total := range coinTotals {
		wins := coinWins[coin]
		if total > 0 {
			result.CoinWinRate[coin] = float64(wins) / float64(total) * 100
		}
		result.CoinPnl[coin] = coinPnls[coin]
	}

	// 计算持仓时间分布
	result.HoldingTimeDistribution = s.calculateHoldingTimeDistribution(holdingTimes)

	return result
}

// calculateFollowProfit 计算跟投收益（模拟真实跟单，带资金曲线和最大回撤）
func (s *TraderAnalysisService) calculateFollowProfit(traderID, timeRange string, initialCapital, investPerOrder float64) *FollowProfitResult {
	orders, err := weex.GetHistoryOrderList(traderID)
	if err != nil {
		s.logger.Error("Failed to get history orders for follow profit calculation", "error", err)
		return nil
	}

	filteredOrders := s.filterOrdersByTimeRange(orders, timeRange)

	result := &FollowProfitResult{
		InitialCapital: initialCapital,
		InvestPerOrder: investPerOrder,
		CapitalCurve:   make([]CapitalDataPoint, 0),
	}

	type eventType int
	// 定义事件类型
	const (
		eventOpen eventType = iota
		eventClose
	)

	// 定义事件结构
	type followEvent struct {
		time      time.Time
		eventType eventType
		profit    float64
	}

	// 创建事件流
	var events []followEvent
	for _, order := range filteredOrders {
		openTimestamp, _ := strconv.ParseInt(order.OpenTime, 10, 64)
		openTime := time.Unix(openTimestamp/1000, 0)
		events = append(events, followEvent{time: openTime, eventType: eventOpen})

		if order.CloseTime != "" {
			closeTimestamp, _ := strconv.ParseInt(order.CloseTime, 10, 64)
			closeTime := time.Unix(closeTimestamp/1000, 0)
			profitRate, _ := strconv.ParseFloat(order.ProfitRate, 64)
			orderProfit := investPerOrder * (profitRate / 100)
			events = append(events, followEvent{time: closeTime, eventType: eventClose, profit: orderProfit})
		}
	}

	// 按时间排序事件
	sort.Slice(events, func(i, j int) bool {
		return events[i].time.Before(events[j].time)
	})

	currentCapital := initialCapital
	peakCapital := initialCapital
	maxDrawdown := 0.0
	ordersFollowed := 0
	activePositions := 0

	// 添加初始点
	if len(events) > 0 {
		result.CapitalCurve = append(result.CapitalCurve, CapitalDataPoint{Time: events[0].time.Add(-time.Second), Capital: initialCapital})
	} else {
		result.CapitalCurve = append(result.CapitalCurve, CapitalDataPoint{Time: time.Now(), Capital: initialCapital})
	}

	// 处理事件流
	for _, e := range events {
		switch e.eventType {
		case eventOpen:
			if currentCapital >= investPerOrder {
				currentCapital -= investPerOrder
				ordersFollowed++
				activePositions++
			}
		case eventClose:
			if activePositions > 0 {
				currentCapital += investPerOrder + e.profit
				activePositions--
			}
		}

		// 更新资金峰值和最大回撤
		if currentCapital > peakCapital {
			peakCapital = currentCapital
		}
		drawdown := (peakCapital - currentCapital) / peakCapital
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}

		// 记录资金曲线数据点
		result.CapitalCurve = append(result.CapitalCurve, CapitalDataPoint{Time: e.time, Capital: currentCapital})
	}

	result.FinalCapital = currentCapital
	result.OrdersFollowed = ordersFollowed
	result.TotalProfit = result.FinalCapital - result.InitialCapital
	result.MaxDrawdown = maxDrawdown * 100 // 转换为百分比
	if result.InitialCapital > 0 {
		result.ProfitRate = (result.TotalProfit / result.InitialCapital) * 100
	}

	return result
}

// 辅助函数
func (s *TraderAnalysisService) max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func (s *TraderAnalysisService) average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (s *TraderAnalysisService) percentile(values []float64, p int) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := float64(p) / 100 * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// func (s *TraderAnalysisService) standardDeviation(values []float64) float64 {
// 	if len(values) <= 1 {
// 		return 0
// 	}

// 	mean := s.average(values)
// 	variance := 0.0
// 	for _, v := range values {
// 		variance += math.Pow(v-mean, 2)
// 	}
// 	variance /= float64(len(values) - 1)
// 	return math.Sqrt(variance)
// }

func (s *TraderAnalysisService) calculateHoldingTimeDistribution(holdingTimes []float64) []HoldingTimeBucket {
	if len(holdingTimes) == 0 {
		return nil
	}

	buckets := []HoldingTimeBucket{
		{Range: "< 1小时", Count: 0},
		{Range: "1-6小时", Count: 0},
		{Range: "6-24小时", Count: 0},
		{Range: "1-3天", Count: 0},
		{Range: "3-7天", Count: 0},
		{Range: "> 7天", Count: 0},
	}

	for _, hours := range holdingTimes {
		switch {
		case hours < 1:
			buckets[0].Count++
		case hours < 6:
			buckets[1].Count++
		case hours < 24:
			buckets[2].Count++
		case hours < 72:
			buckets[3].Count++
		case hours < 168:
			buckets[4].Count++
		default:
			buckets[5].Count++
		}
	}

	total := len(holdingTimes)
	for i := range buckets {
		buckets[i].Rate = float64(buckets[i].Count) / float64(total) * 100
	}

	return buckets
}
