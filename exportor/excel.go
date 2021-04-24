// 导出结果为 excel

package exportor

import (
	"context"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/axiaoxin-com/logging"
)

var (

	// HeaderStyle 表头样式
	HeaderStyle = &excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"FFCCCC"},
			Shading: 0,
		},
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal:      "center",
			JustifyLastLine: true,
			Vertical:        "center",
			WrapText:        true,
		},
	}
	// BodyStyle 表格Style
	BodyStyle = &excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal:      "left",
			JustifyLastLine: true,
			Vertical:        "center",
			WrapText:        true,
		},
	}
)

// ExportExcel 导出 excel
func (e Exportor) ExportExcel(ctx context.Context, filename string) (result []byte, err error) {
	f := excelize.NewFile()

	// 创建全部数据表
	defaultSheet := "Sheet1"
	lowPriceSheet := "30元内"
	hv1Sheet := "波动率低于0.1"
	hv2Sheet := "波动率0.1-0.5"
	hv3Sheet := "波动率高于0.5"
	sheets := []string{defaultSheet, lowPriceSheet, hv1Sheet, hv2Sheet, hv3Sheet}
	// 添加行业
	for _, industry := range e.Stocks.GetIndustryList() {
		sheets = append(sheets, industry+"行业")
	}

	headers := e.Stocks[0].GetHeaders()
	headerStyle, err := f.NewStyle(HeaderStyle)
	if err != nil {
		logging.Error(ctx, "New HeaderStyle error:"+err.Error())
	}
	bodyStyle, err := f.NewStyle(BodyStyle)
	if err != nil {
		logging.Error(ctx, "New BodyStyle error:"+err.Error())
	}
	// 创建 sheet
	for _, sheet := range sheets {
		f.NewSheet(sheet)
		for i, header := range headers {
			// 设置列宽
			colNum := i + 1
			width := 30.0
			switch header {
			case "主营构成", "每股收益预测":
				width = 45.0
			case "公司信息":
				width = 75.0
			}
			col, err := excelize.ColumnNumberToName(colNum)
			if err != nil {
				logging.Error(ctx, "CoordinatesToCellName error:"+err.Error())
			}
			f.SetColWidth(sheet, col, col, width)
			// 设置表头行高
			rowNum := 1
			height := 20.0
			f.SetRowHeight(sheet, rowNum, height)
		}

		// 设置表头样式
		hcell, err := excelize.CoordinatesToCellName(1, 1)
		if err != nil {
			logging.Error(ctx, "CoordinatesToCellName error:"+err.Error())
			continue
		}
		vcell, err := excelize.CoordinatesToCellName(len(headers), 1)
		if err != nil {
			logging.Error(ctx, "CoordinatesToCellName error:"+err.Error())
			continue
		}
		f.SetCellStyle(sheet, hcell, vcell, headerStyle)

		// 设置表格样式
		hcell, err = excelize.CoordinatesToCellName(1, 2)
		if err != nil {
			logging.Error(ctx, "CoordinatesToCellName error:"+err.Error())
			continue
		}
		vcell, err = excelize.CoordinatesToCellName(len(headers), len(e.Stocks))
		if err != nil {
			logging.Error(ctx, "CoordinatesToCellName error:"+err.Error())
			continue
		}
		f.SetCellStyle(sheet, hcell, vcell, bodyStyle)
	}

	// 写 header
	for _, sheet := range sheets {
		for i, header := range headers {
			axis, err := excelize.CoordinatesToCellName(i+1, 1)
			if err != nil {
				logging.Error(ctx, "CoordinatesToCellName error:"+err.Error())
				continue
			}
			f.SetCellValue(sheet, axis, header)
		}
	}
	// 写 body
	e.Stocks.SortByROE()
	for _, sheet := range sheets {
		row := 2
		for _, stock := range e.Stocks {
			switch sheet {
			case defaultSheet:
			case lowPriceSheet:
				if stock.Price > 30 {
					continue
				}
			case hv1Sheet:
				if stock.HV > 0.1 {
					continue
				}
			case hv2Sheet:
				if stock.HV <= 0.1 || stock.HV > 0.5 {
					continue
				}
			case hv3Sheet:
				if stock.HV <= 0.5 {
					continue
				}
			}
			if strings.HasSuffix(sheet, "行业") && !strings.Contains(sheet, stock.Industry) {
				continue
			}
			headerValueMap := stock.GetHeaderValueMap()
			for k, header := range headers {
				col := k + 1
				axis, err := excelize.CoordinatesToCellName(col, row)
				if err != nil {
					logging.Error(ctx, "CoordinatesToCellName error:"+err.Error())
					continue
				}
				value := headerValueMap[header]
				f.SetCellValue(sheet, axis, value)
			}
			row++
		}
	}

	buf, err := f.WriteToBuffer()
	result = buf.Bytes()
	err = f.SaveAs(filename)
	return
}
