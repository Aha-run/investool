// 天天基金获取基金经理列表(web接口)
// https://fund.eastmoney.com/manager/jjjl_all_penavgrowth_desc.html

package eastmoney

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/axiaoxin-com/goutils"
	"github.com/axiaoxin-com/logging"
	"github.com/corpix/uarand"
	"go.uber.org/zap"
)

// FundManagerInfo 基金经理信息
type FundManagerInfo struct {
	// ID
	ID string `json:"id"`
	// 姓名
	Name string `json:"name"`
	// 基金公司id
	FundCompanyID string `json:"fund_company_id"`
	// 基金公司名称
	FundCompanyName string `json:"fund_company_name"`
	// 现任基金代码列表
	FundCodes []string `json:"fund_codes"`
	// 现任基金名称列表
	FundNames []string `json:"fund_names"`
	// 累计从业时间(天)
	WorkingDays int `json:"working_days"`
	// 现任基金最佳回报（%）
	CurrentBestReturn float64 `json:"current_best_return"`
	// 现任最佳基金代码
	CurrentBestFundCode string `json:"current_best_fund_code"`
	// 现任最佳基金名称
	CurrentBestFundName string `json:"current_best_fund_name"`
	// 现任基金资产总规模（亿元）
	CurrentFundScale float64 `json:"current_fund_scale"`
	// 任职期间最佳基金回报
	WorkingBestReturn float64 `json:"working_best_return"`
}

// FundMangers 查询基金经理列表（web接口）
// ft（基金类型） all:全部 gp:股票型 hh:混合型 zq:债券型 sy:收益型
// sc（排序字段）abbname:经理名 jjgspy:基金公司 totaldays:从业时间 netnav:基金规模 penavgrowth:现任基金最佳回报
// st（排序类型）asc desc
func (e EastMoney) FundMangers(ctx context.Context, ft, sc, st string) ([]*FundManagerInfo, error) {
	beginTime := time.Now()
	header := map[string]string{
		"user-agent": uarand.GetRandom(),
	}
	result := []*FundManagerInfo{}
	index := 1
	for {
		apiurl := fmt.Sprintf(
			"http://fund.eastmoney.com/Data/FundDataPortfolio_Interface.aspx?dt=14&mc=returnjson&ft=%s&pn=20&pi=%d&sc=%s&st=%s",
			ft,
			index,
			sc,
			st,
		)
		logging.Debug(ctx, "EastMoney FundMangers "+apiurl+" begin", zap.Int("index", index))
		beginTime := time.Now()
		resp, err := goutils.HTTPGETRaw(ctx, e.HTTPClient, apiurl, header)
		strresp := string(resp)
		latency := time.Now().Sub(beginTime).Milliseconds()
		logging.Debug(ctx, "EastMoney FundMangers "+apiurl+" end",
			zap.Int64("latency(ms)", latency),
			zap.Int("index", index),
		)
		if err != nil {
			return nil, err
		}
		reg, err := regexp.Compile(`\[(".+?")\]`)
		if err != nil {
			logging.Error(ctx, "regexp error:"+err.Error())
			return nil, err
		}
		matched := reg.FindAllStringSubmatch(strresp, -1)
		if len(matched) == 0 {
			break
		}

		for _, m := range matched {
			// "30057445","张少华","80000200","中银证券","009640,009641,010892,010893,501095","中银证券优选行业龙头混合A,中银证券优选行业龙头混合C,中银证券精选行业股票A,中银证券精选行业股票C,中银证券科创3年封闭混合","3993","26.52%","009640","中银证券优选行业龙头混合A","28.43亿元","82.97%"
			field, _ := regexp.Compile(`"(.*?)"`)
			fields := field.FindAllStringSubmatch(m[1], -1)
			if len(fields) != 12 {
				logging.Warnf(ctx, "invalid fields len:%v %v", len(fields), m[1])
				continue
			}
			totaldays := 0
			if fields[6][1] != "" && fields[6][1] != "--" {
				totaldays, err = strconv.Atoi(fields[6][1])
				if err != nil {
					logging.Warnf(ctx, "parse totaldays:%v to int error:%v", fields[6], err)
				}
			}
			bestReturn := 0.0
			if fields[7][1] != "" && fields[7][1] != "--" {
				bestReturnNum := strings.TrimSuffix(fields[7][1], "%")
				bestReturn, err = strconv.ParseFloat(bestReturnNum, 64)
				if err != nil {
					logging.Warnf(ctx, "parse bestReturn:%v to float64 error:%v", bestReturnNum, err)
				}
			}
			scale := 0.0
			if fields[10][1] != "" && fields[10][1] != "--" {
				scaleNum := strings.TrimSuffix(fields[10][1], "亿元")
				scale, err = strconv.ParseFloat(scaleNum, 64)
				if err != nil {
					logging.Warnf(ctx, "parse scale:%v to float64 error:%v", scaleNum, err)
				}
			}
			wbestReturn := 0.0
			if fields[11][1] != "" && fields[11][1] != "--" {
				wbestReturnNum := strings.TrimSuffix(fields[11][1], "%")
				wbestReturn, err = strconv.ParseFloat(wbestReturnNum, 64)
				if err != nil {
					logging.Warnf(ctx, "parse bestReturn:%v to float64 error:%v", wbestReturnNum, err)
				}
			}
			result = append(result, &FundManagerInfo{
				ID:                  fields[0][1],
				Name:                fields[1][1],
				FundCompanyID:       fields[2][1],
				FundCompanyName:     fields[3][1],
				FundCodes:           strings.Split(fields[4][1], ","),
				FundNames:           strings.Split(fields[5][1], ","),
				WorkingDays:         totaldays,
				CurrentBestReturn:   bestReturn,
				CurrentBestFundCode: fields[8][1],
				CurrentBestFundName: fields[9][1],
				CurrentFundScale:    scale,
				WorkingBestReturn:   wbestReturn,
			})
		}
		index++
	}
	latency := time.Now().Sub(beginTime).Milliseconds()
	logging.Debug(
		ctx,
		"EastMoney FundMangers end",
		zap.Int64("latency(ms)", latency),
		zap.Int("resultCount", len(result)),
	)
	return result, nil
}
