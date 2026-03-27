package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"go-migration-platform/services/education/internal/model"
)

type orderReceiptPDFView struct {
	Template           string
	OrgName            string
	OrderNumber        string
	StudentName        string
	StudentPhone       string
	OrderTypeText      string
	OrderStatusText    string
	OrderSourceText    string
	DealDateText       string
	CreatedTimeText    string
	FinishedTimeText   string
	StaffName          string
	SalePersonName     string
	PrintedBy          string
	PrintedAtText      string
	OrderTagText       string
	Remark             string
	ExternalRemark     string
	AmountInWords      string
	TotalAmountText    string
	DiscountAmountText string
	StorageAmountText  string
	PaidAmountText     string
	ArrearAmountText   string
	PaymentSummaryText string
	CampusPhone        string
	CampusAddress      string
	ItemRows           []orderReceiptPDFItemRow
	PaymentRows        []orderReceiptPDFPaymentRow
}

type orderReceiptPDFItemRow struct {
	Name               string
	QuoteLabel         string
	ChargingModeText   string
	QuantityText       string
	GiftText           string
	PeriodText         string
	OriginalAmountText string
	DiscountAmountText string
	AmountText         string
	NoteText           string
}

type orderReceiptPDFPaymentRow struct {
	MethodText  string
	AccountText string
	PaidAtText  string
	AmountText  string
	RemarkText  string
}

var orderReceiptPDFTemplate = template.Must(template.New("order_receipt_pdf").Parse(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>{{.OrderNumber}}</title>
  <style>
    {{if eq .Template "a4"}}
    @page { size: A4; margin: 18mm 10mm 12mm; }
    body { margin: 0; font-family: "PingFang SC","PingFangSC","Microsoft YaHei",sans-serif; color: #222; }
    .sheet { padding-top: 6mm; }
    .org-name { text-align: center; font-size: 28px; font-weight: 700; line-height: 1.2; }
    .divider { border-top: 2px solid #222; margin: 8px 0 12px; }
    .meta { display: grid; grid-template-columns: 1fr 1fr 1.5fr; gap: 12px; font-size: 13px; font-weight: 600; margin-bottom: 12px; }
    .section { margin-top: 12px; }
    .section-title { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; font-size: 14px; font-weight: 700; }
    .section-title .bar { width: 4px; height: 18px; border-radius: 999px; background: #1677ff; }
    table { width: 100%; border-collapse: collapse; table-layout: fixed; font-size: 12px; }
    th, td { border: 1px solid #808a98; padding: 9px 10px; line-height: 1.4; vertical-align: middle; text-align: left; word-break: break-word; }
    th { font-weight: 700; }
    tbody + tbody tr:first-child td { border-top-width: 2px; }
    .signature { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin-top: 14px; font-size: 13px; }
    {{else if eq .Template "dot"}}
    @page { size: 210mm 150mm; margin: 0; }
    body { margin: 0; font-family: "Courier New","PingFang SC","Microsoft YaHei",monospace; color: #222; }
    .sheet { padding: 14mm 12mm; }
    .org-name { text-align: center; font-size: 22px; font-weight: 700; }
    .sub-title { text-align: center; margin-top: 6px; font-size: 16px; font-weight: 700; }
    .meta-grid, .info-grid, .summary-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px 24px; margin-top: 10px; font-size: 13px; }
    .meta-item, .info-item { display: grid; grid-template-columns: 88px minmax(0, 1fr); gap: 6px; align-items: baseline; }
    .meta-label, .info-label { font-weight: 700; }
    .line { border-top: 1px dashed #8a94a6; margin: 12px 0; }
    .section-title { margin-bottom: 10px; font-weight: 700; }
    .item { padding: 8px 0; border-bottom: 1px dotted #c7cfdd; }
    .item:last-child { border-bottom: none; }
    .item-top, .row { display: flex; justify-content: space-between; gap: 12px; }
    .item-top { font-weight: 700; }
    .sub, .empty { margin-top: 4px; color: #5d6777; font-size: 13px; line-height: 1.5; }
    .footer { display: flex; justify-content: space-between; gap: 14px; margin-top: 10px; font-size: 13px; }
    {{else}}
    @page { size: 80mm 210mm; margin: 0; }
    body { margin: 0; font-family: "Courier New","PingFang SC","Microsoft YaHei",monospace; color: #222; }
    .sheet { padding: 8mm 6mm 9mm; }
    .org-name { text-align: center; font-size: 16px; font-weight: 700; line-height: 1.5; }
    .sub-title { text-align: center; margin-top: 4px; font-size: 14px; font-weight: 700; }
    .sub-meta { margin-top: 4px; text-align: center; font-size: 11px; word-break: break-all; }
    .block, .item { display: flex; flex-direction: column; gap: 4px; }
    .item + .item { margin-top: 8px; }
    .row { display: flex; justify-content: space-between; gap: 12px; font-size: 12px; line-height: 1.5; }
    .sub, .footer { font-size: 11px; color: #6a7380; line-height: 1.5; }
    .line { border-top: 1px dashed #8a94a6; margin: 8px 0; }
    .section-title { margin-bottom: 10px; font-weight: 700; }
    .footer { margin-top: 10px; text-align: right; }
    {{end}}
  </style>
</head>
<body>
  {{if eq .Template "a4"}}
  <div class="sheet">
    <div class="org-name">{{.OrgName}}</div>
    <div class="divider"></div>
    <div class="meta">
      <span>学员姓名：{{.StudentName}}</span>
      <span>订单类型：{{.OrderTypeText}}</span>
      <span>订单号：{{.OrderNumber}}</span>
    </div>

    <div class="section">
      <div class="section-title"><span class="bar"></span>校区信息</div>
      <table>
        <tr><th>校区名称</th><th>校区电话</th><th>校区地址</th></tr>
        <tr><td>{{.OrgName}}</td><td>{{.CampusPhone}}</td><td>{{.CampusAddress}}</td></tr>
      </table>
    </div>

    <div class="section">
      <div class="section-title"><span class="bar"></span>商品信息</div>
      <table>
        {{range .ItemRows}}
        <tbody>
          <tr>
            <td colspan="2"><strong>课程名称：</strong>{{.Name}}</td>
            <td colspan="2"><strong>收费方式：</strong>{{.ChargingModeText}}</td>
            <td colspan="2"><strong>报价单：</strong>{{.QuoteLabel}}</td>
          </tr>
          <tr>
            <th>购买数量</th><th>赠送数量</th><th>有效期</th><th>原价</th><th>优惠</th><th>应收小计</th>
          </tr>
          <tr>
            <td>{{.QuantityText}}</td><td>{{.GiftText}}</td><td>{{.PeriodText}}</td><td>{{.OriginalAmountText}}</td><td>{{.DiscountAmountText}}</td><td>{{.AmountText}}</td>
          </tr>
          <tr><td colspan="6"><strong>说明：</strong>{{.NoteText}}</td></tr>
        </tbody>
        {{end}}
      </table>
    </div>

    <div class="section">
      <div class="section-title"><span class="bar"></span>订单信息</div>
      <table>
        <tr><th>经办日期</th><th>经办人</th><th>订单销售员</th><th>订单来源</th></tr>
        <tr><td>{{.DealDateText}}</td><td>{{.StaffName}}</td><td>{{.SalePersonName}}</td><td>{{.OrderSourceText}}</td></tr>
        <tr><th>订单状态</th><th>订单标签</th><th>打印人</th><th>打印时间</th></tr>
        <tr><td>{{.OrderStatusText}}</td><td>{{.OrderTagText}}</td><td>{{.PrintedBy}}</td><td>{{.PrintedAtText}}</td></tr>
      </table>
    </div>

    <div class="section">
      <div class="section-title"><span class="bar"></span>结算信息</div>
      <table>
        <tr><th>订单应收</th><th>整单优惠</th><th>储值抵扣</th><th>外部实收</th><th>待收欠费</th><th>金额大写</th></tr>
        <tr><td>{{.TotalAmountText}}</td><td>{{.DiscountAmountText}}</td><td>{{.StorageAmountText}}</td><td>{{.PaidAmountText}}</td><td>{{.ArrearAmountText}}</td><td>{{.AmountInWords}}</td></tr>
        <tr><th colspan="2">支付摘要</th><td colspan="4">{{.PaymentSummaryText}}</td></tr>
      </table>
    </div>

    <div class="section">
      <div class="section-title"><span class="bar"></span>备注信息</div>
      <table>
        <tr><th>对内备注</th><th>对外备注</th></tr>
        <tr><td>{{.Remark}}</td><td>{{.ExternalRemark}}</td></tr>
      </table>
    </div>

    <div class="signature">
      <div>经办人：{{.StaffName}}</div>
      <div>复核人：________________</div>
      <div>付款人/家长签字：________________</div>
    </div>
  </div>
  {{else if eq .Template "dot"}}
  <div class="sheet">
    <div class="org-name">{{.OrgName}}</div>
    <div class="sub-title">业务收据（针式打印联）</div>
    <div class="meta-grid">
      <div class="meta-item"><span class="meta-label">票据编号：</span><span>{{.OrderNumber}}</span></div>
      <div class="meta-item"><span class="meta-label">打印时间：</span><span>{{.PrintedAtText}}</span></div>
    </div>
    <div class="info-grid">
      <div class="info-item"><span class="info-label">学员：</span><span>{{.StudentName}}</span></div>
      <div class="info-item"><span class="info-label">手机：</span><span>{{.StudentPhone}}</span></div>
      <div class="info-item"><span class="info-label">类型：</span><span>{{.OrderTypeText}}</span></div>
      <div class="info-item"><span class="info-label">状态：</span><span>{{.OrderStatusText}}</span></div>
      <div class="info-item"><span class="info-label">来源：</span><span>{{.OrderSourceText}}</span></div>
      <div class="info-item"><span class="info-label">经办日期：</span><span>{{.DealDateText}}</span></div>
      <div class="info-item"><span class="info-label">经办人：</span><span>{{.StaffName}}</span></div>
      <div class="info-item"><span class="info-label">销售员：</span><span>{{.SalePersonName}}</span></div>
    </div>
    <div class="line"></div>
    <div class="section-title">业务明细</div>
    {{range .ItemRows}}
    <div class="item">
      <div class="item-top"><span>{{.Name}}</span><span>{{.AmountText}}</span></div>
      <div class="sub">{{.QuoteLabel}}｜购买 {{.QuantityText}}｜赠送 {{.GiftText}}</div>
      <div class="sub">{{.PeriodText}}</div>
    </div>
    {{end}}
    <div class="line"></div>
    <div class="section-title">结算信息</div>
    <div class="summary-grid">
      <div>订单应收：{{.TotalAmountText}}</div>
      <div>整单优惠：{{.DiscountAmountText}}</div>
      <div>储值抵扣：{{.StorageAmountText}}</div>
      <div>外部实收：{{.PaidAmountText}}</div>
      <div>待收欠费：{{.ArrearAmountText}}</div>
      <div>金额大写：{{.AmountInWords}}</div>
    </div>
    <div class="line"></div>
    <div class="section-title">支付摘要</div>
    <div class="empty">{{.PaymentSummaryText}}</div>
    <div class="line"></div>
    <div class="footer"><span>经办人：{{.StaffName}}</span><span>打印人：{{.PrintedBy}}</span><span>签字：____________</span></div>
  </div>
  {{else}}
  <div class="sheet">
    <div class="org-name">{{.OrgName}}</div>
    <div class="sub-title">业务收据</div>
    <div class="sub-meta">{{.OrderNumber}}</div>
    <div class="block">
      <div class="row"><span>学员</span><span>{{.StudentName}}</span></div>
      <div class="row"><span>手机</span><span>{{.StudentPhone}}</span></div>
      <div class="row"><span>类型</span><span>{{.OrderTypeText}}</span></div>
      <div class="row"><span>日期</span><span>{{.DealDateText}}</span></div>
      <div class="row"><span>经办人</span><span>{{.StaffName}}</span></div>
    </div>
    <div class="line"></div>
    <div class="section-title">业务明细</div>
    {{range .ItemRows}}
    <div class="item">
      <div class="row"><span>{{.Name}}</span><span>{{.AmountText}}</span></div>
      <div class="sub">{{.QuoteLabel}}</div>
      <div class="sub">购买 {{.QuantityText}} / 赠送 {{.GiftText}}</div>
    </div>
    {{end}}
    <div class="line"></div>
    <div class="section-title">金额汇总</div>
    <div class="row"><span>订单应收</span><span>{{.TotalAmountText}}</span></div>
    <div class="row"><span>整单优惠</span><span>{{.DiscountAmountText}}</span></div>
    <div class="row"><span>储值抵扣</span><span>{{.StorageAmountText}}</span></div>
    <div class="row"><span>外部实收</span><span>{{.PaidAmountText}}</span></div>
    <div class="row"><span>待收欠费</span><span>{{.ArrearAmountText}}</span></div>
    <div class="line"></div>
    <div class="section-title">支付摘要</div>
    <div class="sub">{{.PaymentSummaryText}}</div>
    <div class="sub">金额大写：{{.AmountInWords}}</div>
    <div class="footer">打印人：{{.PrintedBy}}</div>
  </div>
  {{end}}
</body>
</html>
`))

func (svc *Service) GenerateOrderReceiptPDF(userID int64, orderIDRaw, templateRaw string) (string, []byte, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, errors.New("no institution context")
		}
		return "", nil, err
	}

	orderID, err := strconv.ParseInt(strings.TrimSpace(orderIDRaw), 10, 64)
	if err != nil || orderID <= 0 {
		return "", nil, errors.New("订单ID不能为空")
	}

	orderDetail, err := svc.repo.GetOrderDetail(context.Background(), instID, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, errors.New("订单不存在")
		}
		return "", nil, err
	}

	orgName, err := svc.repo.GetInstitutionName(context.Background(), instID)
	if err != nil {
		orgName = ""
	}
	orgName = strings.TrimSpace(orgName)
	if orgName == "" {
		orgName = "总校区"
	}

	templateMode := normalizeReceiptTemplate(templateRaw)
	view := buildOrderReceiptPDFView(orgName, orderDetail, templateMode)
	htmlBytes, err := renderOrderReceiptPDFHTML(view)
	if err != nil {
		return "", nil, err
	}
	pdfBytes, err := generateOrderReceiptPDFBytes(htmlBytes, templateMode)
	if err != nil {
		return "", nil, err
	}

	filename := sanitizeTemplateFileName(fmt.Sprintf("%s-收据-%s-%s.pdf", orgName, orderDetail.OrderNumber, time.Now().Format("20060102150405")))
	return filename, pdfBytes, nil
}

func buildOrderReceiptPDFView(orgName string, detail model.OrderDetailVO, templateMode string) orderReceiptPDFView {
	view := orderReceiptPDFView{
		Template:           templateMode,
		OrgName:            orgName,
		OrderNumber:        fallbackString(detail.OrderNumber, "-"),
		StudentName:        fallbackString(detail.StudentName, "-"),
		StudentPhone:       fallbackString(detail.StudentPhone, "-"),
		OrderTypeText:      orderTypeText(detail.OrderType),
		OrderStatusText:    orderStatusText(detail.OrderStatus),
		OrderSourceText:    orderSourceText(detail.OrderSource),
		DealDateText:       formatDateOnly(detail.DealDate),
		CreatedTimeText:    formatDateTimeValueLocal(detail.CreatedTime),
		FinishedTimeText:   formatDateTime(detail.FinishedTime),
		StaffName:          fallbackString(detail.StaffName, "-"),
		SalePersonName:     fallbackString(detail.SalePersonName, "-"),
		PrintedBy:          fallbackString(detail.StaffName, "-"),
		PrintedAtText:      time.Now().Format("2006-01-02 15:04:05"),
		OrderTagText:       joinWithFallback(detail.OrderTagNames, "、", "-"),
		Remark:             fallbackString(detail.Remark, "-"),
		ExternalRemark:     fallbackString(detail.ExternalRemark, "-"),
		AmountInWords:      formatChineseMoney(detail.TotalAmount),
		TotalAmountText:    formatCurrency(detail.TotalAmount),
		DiscountAmountText: formatSignedDeduction(detail.OrderDiscountAmount),
		StorageAmountText:  formatSignedDeduction(detail.RechargeAccountAmount + detail.RechargeAccountResidualAmount + detail.RechargeAccountGivingAmount),
		PaidAmountText:     formatCurrency(detail.PaidAmount),
		ArrearAmountText:   formatCurrency(detail.ArrearAmount),
		PaymentSummaryText: buildOrderReceiptPaymentSummary(detail.PaymentRecords),
		CampusPhone:        "-",
		CampusAddress:      "-",
	}

	view.ItemRows = buildOrderReceiptPDFItemRows(detail)
	view.PaymentRows = buildOrderReceiptPDFPaymentRows(detail.PaymentRecords)
	return view
}

func buildOrderReceiptPDFItemRows(detail model.OrderDetailVO) []orderReceiptPDFItemRow {
	orderType := derefInt(detail.OrderType)
	if orderType == model.OrderTypeRechargeAccount || orderType == model.OrderTypeRechargeAccountRefund {
		rows := make([]orderReceiptPDFItemRow, 0, 3)
		appendRow := func(name string, amount float64) {
			if amount <= 0 {
				return
			}
			rows = append(rows, orderReceiptPDFItemRow{
				Name:               name,
				QuoteLabel:         map[bool]string{true: "储值账户退费", false: "储值账户充值"}[orderType == model.OrderTypeRechargeAccountRefund],
				ChargingModeText:   "账户金额",
				QuantityText:       "1笔",
				GiftText:           "-",
				PeriodText:         "即时生效",
				OriginalAmountText: formatCurrency(amount),
				DiscountAmountText: "¥0.00",
				AmountText:         formatCurrency(amount),
				NoteText:           "业务类型：" + orderTypeText(detail.OrderType),
			})
		}
		appendRow("充值金额", detail.RechargeAccountAmount)
		appendRow("残联金额", detail.RechargeAccountResidualAmount)
		appendRow("赠送金额", detail.RechargeAccountGivingAmount)
		return rows
	}

	rows := make([]orderReceiptPDFItemRow, 0, len(detail.OrderItems))
	for _, item := range detail.OrderItems {
		chargingMode := derefInt(item.ChargingMode)
		purchaseQuantity := getOrderReceiptPurchaseQuantity(item)
		discountAmount := item.Amount - item.ReceivableAmount
		if discountAmount < 0 {
			discountAmount = 0
		}
		noteParts := make([]string, 0, 3)
		if strings.TrimSpace(item.QuoteName) != "" {
			noteParts = append(noteParts, "报价单："+item.QuoteName)
		}
		if mode := chargingModeText(item.ChargingMode); mode != "-" {
			noteParts = append(noteParts, "收费方式："+mode)
		}
		if item.QuotePrice > 0 {
			noteParts = append(noteParts, "报价："+formatCurrency(item.QuotePrice))
		}
		rows = append(rows, orderReceiptPDFItemRow{
			Name:               fallbackString(item.CourseName, "-"),
			QuoteLabel:         fallbackString(item.QuoteName, "-"),
			ChargingModeText:   chargingModeText(item.ChargingMode),
			QuantityText:       formatQuantity(purchaseQuantity, chargingMode),
			GiftText:           formatGiftText(item.FreeQuantity, chargingMode),
			PeriodText:         buildOrderReceiptPeriodText(item.ValidDate, item.EndDate),
			OriginalAmountText: formatCurrency(item.Amount),
			DiscountAmountText: formatSignedDeduction(discountAmount),
			AmountText:         formatCurrency(item.ReceivableAmount),
			NoteText:           fallbackString(strings.Join(noteParts, "｜"), "-"),
		})
	}
	if len(rows) == 0 {
		rows = append(rows, orderReceiptPDFItemRow{
			Name:               "暂无明细",
			QuoteLabel:         "-",
			ChargingModeText:   "-",
			QuantityText:       "-",
			GiftText:           "-",
			PeriodText:         "-",
			OriginalAmountText: "¥0.00",
			DiscountAmountText: "¥0.00",
			AmountText:         "¥0.00",
			NoteText:           "-",
		})
	}
	return rows
}

func buildOrderReceiptPDFPaymentRows(records []model.OrderPaymentRecordVO) []orderReceiptPDFPaymentRow {
	rows := make([]orderReceiptPDFPaymentRow, 0, len(records))
	for _, item := range records {
		if item.PayAmount <= 0 {
			continue
		}
		rows = append(rows, orderReceiptPDFPaymentRow{
			MethodText:  payMethodText(item.PayMethod),
			AccountText: fallbackString(item.AccountName, "-"),
			PaidAtText:  formatDateTime(coalesceTime(item.PayTime, item.CreatedTime)),
			AmountText:  formatCurrency(item.PayAmount),
			RemarkText:  fallbackString(item.Remark, "-"),
		})
	}
	return rows
}

func renderOrderReceiptPDFHTML(view orderReceiptPDFView) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := orderReceiptPDFTemplate.Execute(buf, view); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func generateOrderReceiptPDFBytes(htmlBytes []byte, templateMode string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoFirstRun,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("font-render-hinting", "medium"),
	)
	if execPath := findChromeExecPath(); execPath != "" {
		opts = append(opts, chromedp.ExecPath(execPath))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()
	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	var pdfBytes []byte
	widthInch, heightInch := orderReceiptPaperSizeInch(templateMode)
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, string(htmlBytes)).Do(ctx)
		}),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				WithPaperWidth(widthInch).
				WithPaperHeight(heightInch).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBytes = buf
			return nil
		}),
	); err != nil {
		return nil, err
	}

	return pdfBytes, nil
}

func normalizeReceiptTemplate(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "dot":
		return "dot"
	case "receipt":
		return "receipt"
	default:
		return "a4"
	}
}

func orderReceiptPaperSizeInch(templateMode string) (float64, float64) {
	switch templateMode {
	case "dot":
		return 210.0 / 25.4, 150.0 / 25.4
	case "receipt":
		return 80.0 / 25.4, 210.0 / 25.4
	default:
		return 210.0 / 25.4, 297.0 / 25.4
	}
}

func findChromeExecPath() string {
	candidates := []string{
		os.Getenv("CHROME_PATH"),
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/opt/homebrew/bin/google-chrome",
		"/opt/homebrew/bin/chromium",
	}
	for _, path := range candidates {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func orderTypeText(orderType *int) string {
	switch derefInt(orderType) {
	case model.OrderTypeRegistrationRenewal:
		return "报名续费"
	case model.OrderTypeRechargeAccount:
		return "储值账户充值"
	case model.OrderTypeRefundCourse:
		return "退课"
	case model.OrderTypeRechargeAccountRefund:
		return "储值账户退费"
	case model.OrderTypeTransferCourse:
		return "转课"
	case model.OrderTypeRefundMaterialFee:
		return "退教材费"
	case model.OrderTypeRefundMiscFee:
		return "退学杂费"
	default:
		return "-"
	}
}

func orderStatusText(status *int) string {
	switch derefInt(status) {
	case model.OrderStatusPendingPayment:
		return "待付款"
	case model.OrderStatusApproving:
		return "审批中"
	case model.OrderStatusCompleted:
		return "已完成"
	case model.OrderStatusClosed:
		return "已关闭"
	case model.OrderStatusVoided:
		return "已作废"
	case model.OrderStatusPendingHandle:
		return "待处理"
	case model.OrderStatusRefunding:
		return "退费中"
	case model.OrderStatusRefunded:
		return "已退费"
	default:
		return "-"
	}
}

func orderSourceText(source *int) string {
	switch derefInt(source) {
	case model.OrderSourceOffline:
		return "线下办理"
	case model.OrderSourceMiniProgram:
		return "微校报名"
	case model.OrderSourceOfflineImport:
		return "线下导入"
	case model.OrderSourceRenewalOrder:
		return "续费订单"
	default:
		return "-"
	}
}

func payMethodText(payMethod *int) string {
	switch derefInt(payMethod) {
	case 1:
		return "微信"
	case 2:
		return "支付宝"
	case 3:
		return "银行转账"
	case 4:
		return "POS机"
	case 5:
		return "现金"
	case 6:
		return "其他"
	default:
		return "-"
	}
}

func chargingModeText(mode *int) string {
	switch derefInt(mode) {
	case 1:
		return "按课时"
	case 2:
		return "按时段"
	case 3:
		return "按金额"
	default:
		return "-"
	}
}

func buildOrderReceiptPaymentSummary(records []model.OrderPaymentRecordVO) string {
	items := make([]string, 0, len(records))
	for _, item := range records {
		if item.PayAmount <= 0 {
			continue
		}
		items = append(items, fmt.Sprintf("%s %s", payMethodText(item.PayMethod), formatCurrency(item.PayAmount)))
	}
	if len(items) == 0 {
		return "暂无支付明细"
	}
	return strings.Join(items, "；")
}

func buildOrderReceiptPeriodText(validDate, endDate *time.Time) string {
	start := formatDateOnly(validDate)
	end := formatDateOnly(endDate)
	if start != "-" && end != "-" {
		return start + " 至 " + end
	}
	if end != "-" {
		return "有效期至 " + end
	}
	if start != "-" {
		return "开始于 " + start
	}
	return "不限期"
}

func getOrderReceiptPurchaseQuantity(item model.OrderCourseDetailVO) float64 {
	if item.RealQuantity > 0 {
		value := item.RealQuantity - item.FreeQuantity
		if value < 0 {
			return 0
		}
		return value
	}
	return item.Count
}

func formatQuantity(value float64, chargingMode int) string {
	text := formatNumber(value)
	switch chargingMode {
	case 1:
		return text + "课时"
	case 2:
		return text + "天"
	case 3:
		return text + "元"
	default:
		return text
	}
}

func formatGiftText(value float64, chargingMode int) string {
	if value <= 0 {
		return "-"
	}
	return formatQuantity(value, chargingMode)
}

func formatNumber(value float64) string {
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func formatCurrency(value float64) string {
	return fmt.Sprintf("¥%.2f", value)
}

func formatSignedDeduction(value float64) string {
	if value > 0 {
		return fmt.Sprintf("-¥%.2f", value)
	}
	return "¥0.00"
}

func formatDateOnly(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "-"
	}
	return value.Format("2006-01-02")
}

func formatDateTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "-"
	}
	return value.Format("2006-01-02 15:04:05")
}

func formatDateTimeValueLocal(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Format("2006-01-02 15:04:05")
}

func coalesceTime(values ...*time.Time) *time.Time {
	for _, value := range values {
		if value != nil && !value.IsZero() {
			return value
		}
	}
	return nil
}

func fallbackString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func joinWithFallback(items []string, sep, fallback string) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			values = append(values, item)
		}
	}
	if len(values) == 0 {
		return fallback
	}
	return strings.Join(values, sep)
}

func derefInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func formatChineseMoney(value float64) string {
	amount := value
	if amount < 0 {
		amount = -amount
	}
	if amount == 0 {
		return "零元整"
	}
	cnNums := []string{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"}
	cnIntRadice := []string{"", "拾", "佰", "仟"}
	cnIntUnits := []string{"", "万", "亿", "兆"}
	cnDecUnits := []string{"角", "分"}

	rounded := mathRound(amount*100) / 100
	integerNum := int64(rounded)
	decimalNum := int(mathRound((rounded - float64(integerNum)) * 100))

	result := ""
	if integerNum > 0 {
		digits := strings.Split(reverseString(strconv.FormatInt(integerNum, 10)), "")
		zeroCount := 0
		for i, raw := range digits {
			if raw == "" {
				continue
			}
			digit := int(raw[0] - '0')
			quotient := i / 4
			modulus := i % 4
			if digit == 0 {
				zeroCount++
			} else {
				if zeroCount > 0 {
					result = cnNums[0] + result
				}
				zeroCount = 0
				result = cnNums[digit] + cnIntRadice[modulus] + result
			}
			if modulus == 0 && zeroCount < 4 {
				result = cnIntUnits[quotient] + result
			}
		}
		result += "元"
	}
	if decimalNum == 0 {
		result += "整"
	} else {
		jiao := decimalNum / 10
		fen := decimalNum % 10
		if jiao > 0 {
			result += cnNums[jiao] + cnDecUnits[0]
		}
		if fen > 0 {
			result += cnNums[fen] + cnDecUnits[1]
		}
	}
	return result
}

func reverseString(value string) string {
	runes := []rune(value)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func mathRound(value float64) float64 {
	if value < 0 {
		return float64(int64(value - 0.5))
	}
	return float64(int64(value + 0.5))
}
