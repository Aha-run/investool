// 获取历史股价

package eniu

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/axiaoxin-com/goutils"
	"github.com/axiaoxin-com/logging"
	"go.uber.org/zap"
)

// RespHistoricalStockPrice 历史股价接口返回结构
type RespHistoricalStockPrice struct {
	Date  []string  `json:"date"`
	Price []float64 `json:"price"`
}

// HistoricalVolatility 计算历史波动率
// 历史波动率计算方法：https://goodcalculators.com/historical-volatility-calculator/
// 1、从市场上获得标的股票在固定时间间隔(如每天DAY、每周WEEK或每月MONTH等)上的价格。
// 2、对于每个时间段，求出该时间段末的股价与该时段初的股价之比的自然对数。
// 3、求出这些对数值的标准差即为历史波动率的估计值
// 4、若将日、周等标准差转化为年标准差，需要再乘以一年中包含的时段数量的平方根(如，选取时间间隔为每天，则若扣除闭市，每年中有250个交易日，应乘以根号250)
func (p RespHistoricalStockPrice) HistoricalVolatility(ctx context.Context, interval string) (float64, error) {
	// 求末初股价比自然对数
	logs := []float64{}
	for i := len(p.Price) - 1; i >= 1; i-- {
		endPrice := p.Price[i]
		startPrice := p.Price[i-1]
		log := math.Log(endPrice / startPrice)
		logs = append(logs, log)
	}
	// 标准差
	stdev, err := goutils.StdDeviationFloat64(logs)
	if err != nil {
		return 0, err
	}
	logging.Debugs(ctx, "stdev:", stdev)

	intervalValue := float64(250)
	interval = strings.ToUpper(interval)
	switch interval {
	case "DAY":
		intervalValue = 1
	case "WEEK":
		intervalValue = 5
	case "MONTH":
		intervalValue = 21.75
	case "YEAR":
		intervalValue = 250
	}
	volatility := stdev * math.Sqrt(intervalValue)
	return volatility, nil
}

// QueryHistoricalStockPrice 获取历史股价
func (e Eniu) QueryHistoricalStockPrice(ctx context.Context, secuCode string) (RespHistoricalStockPrice, error) {
	apiurl := fmt.Sprintf("https://eniu.com/chart/pricea/%s/t/all", e.GetPathCode(ctx, secuCode))
	logging.Debug(ctx, "EastMoney QueryOrgRating "+apiurl+" begin")
	beginTime := time.Now()
	resp := RespHistoricalStockPrice{}
	err := goutils.HTTPGET(ctx, e.HTTPClient, apiurl, &resp)
	latency := time.Now().Sub(beginTime).Milliseconds()
	logging.Debug(ctx, "EastMoney QueryOrgRating "+apiurl+" end", zap.Int64("latency(ms)", latency), zap.Any("resp", resp))
	return resp, err
}

// GetPathCode 返回接口 url path 中的股票代码
func (e Eniu) GetPathCode(ctx context.Context, secuCode string) string {
	s := strings.Split(secuCode, ".")
	if len(s) != 2 {
		return ""
	}
	return strings.ToLower(s[1]) + s[0]
}
