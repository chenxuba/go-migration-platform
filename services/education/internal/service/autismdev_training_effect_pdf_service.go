package service

import (
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"sort"
	"strings"
	"sync"

	"github.com/signintech/gopdf"
	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/model"
)

const (
	autismDevTrainingEffectTemplateDir     = "assets/autismdev_training_effect"
	autismDevTrainingEffectScoreFontSize   = 15.0
	autismDevTrainingEffectCheckCellWidth  = 18.0
	autismDevTrainingEffectCheckCellHeight = 18.0
)

//go:embed assets/autismdev_training_effect/*.png
var autismDevTrainingEffectTemplateImages embed.FS

type autismDevTrainingEffectRecord struct {
	Record model.AssessmentRecordDetailVO
	Scores map[int]string
}

type autismDevTrainingEffectRow struct {
	Item              autismdevscore.ItemDefinition
	FirstScore        string
	SecondScore       string
	FirstEffect       string
	SecondEffect      string
	SecondEffectReady bool
}

type autismDevTrainingInputScope struct {
	ScopeMode           string   `json:"scopeMode,omitempty"`
	AssessmentScopeMode string   `json:"assessmentScopeMode,omitempty"`
	ScopeDomainCodes    []string `json:"scopeDomainCodes,omitempty"`
	SelectedDomainCodes []string `json:"selectedDomainCodes,omitempty"`
}

type autismDevTrainingEffectTemplatePage struct {
	PageNo          int
	DomainCode      string
	FirstItemNo     int
	LastItemNo      int
	TrailingHeaders int
	Continuation    bool
}

type autismDevTrainingEffectTemplateLayout struct {
	PageNo       int
	Width        int
	Height       int
	ColumnXs     []float64
	RowBounds    []float64
	FirstItemNo  int
	LastItemNo   int
	Continuation bool
}

var (
	autismDevTrainingEffectTemplateLayoutMu    sync.RWMutex
	autismDevTrainingEffectTemplateLayoutCache = map[int]autismDevTrainingEffectTemplateLayout{}
)

func buildAutismDevTrainingEffectPDF(current model.AssessmentRecordDetailVO, records []model.AssessmentRecordDetailVO, institutionName string) ([]byte, error) {
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: autismDevProfilePDFPageWidth, H: autismDevProfilePDFPageHeight},
	})
	if err := addAutismDevProfilePDFFont(&pdf); err != nil {
		return nil, err
	}
	if err := drawAutismDevTrainingEffectPDFPages(&pdf, current, records, institutionName); err != nil {
		return nil, err
	}
	return pdf.GetBytesPdfReturnErr()
}

func drawAutismDevTrainingEffectPDFPages(pdf *gopdf.GoPdf, current model.AssessmentRecordDetailVO, records []model.AssessmentRecordDetailVO, institutionName string) error {
	_ = institutionName

	data, err := loadAutismDevStaticData()
	if err != nil {
		return err
	}

	orderedDomains := autismDevTrainingCurrentRecordDomains(current, data)
	if len(orderedDomains) == 0 {
		return fmt.Errorf("当前记录没有可生成训练效果表的测评领域")
	}

	trainingRecords, err := autismDevTrainingEffectRecords(records, current)
	if err != nil {
		return err
	}
	rowsByDomain := autismDevTrainingEffectRowsByDomain(data.items, orderedDomains, trainingRecords)
	if len(rowsByDomain) == 0 {
		return fmt.Errorf("当前记录没有可生成训练效果表的测评领域")
	}

	renderer := autismDevTrainingEffectPDFRenderer{
		pdf: pdf,
	}
	for _, domainCode := range orderedDomains {
		rows := rowsByDomain[domainCode]
		if len(rows) == 0 {
			continue
		}
		if err := renderer.drawDomain(domainCode, autismDevTrainingDomainName(domainCode, data), rows); err != nil {
			return err
		}
	}
	return nil
}

type autismDevTrainingEffectPDFRenderer struct {
	pdf    *gopdf.GoPdf
	pageNo int
}

func (r *autismDevTrainingEffectPDFRenderer) drawDomain(domainCode, domainName string, rows []autismDevTrainingEffectRow) error {
	rowsByDomainItemNo := make(map[int]autismDevTrainingEffectRow, len(rows))
	for _, row := range rows {
		rowsByDomainItemNo[row.Item.DomainItemNo] = row
	}
	for _, templatePage := range autismDevTrainingEffectTemplatePagesForDomain(domainCode) {
		if err := r.drawTemplatePage(templatePage, domainName, rowsByDomainItemNo); err != nil {
			return err
		}
	}
	return nil
}

func (r *autismDevTrainingEffectPDFRenderer) drawTemplatePage(templatePage autismDevTrainingEffectTemplatePage, domainName string, rows map[int]autismDevTrainingEffectRow) error {
	r.pdf.AddPage()
	r.pageNo++
	if err := r.drawTemplateBackground(templatePage.PageNo); err != nil {
		return err
	}
	layout, err := autismDevTrainingEffectTemplateLayoutForPage(templatePage)
	if err != nil {
		return err
	}
	for itemNo := templatePage.FirstItemNo; itemNo <= templatePage.LastItemNo; itemNo++ {
		row, ok := rows[itemNo]
		if !ok {
			continue
		}
		itemIndex := itemNo - templatePage.FirstItemNo
		if itemIndex < 0 || itemIndex+1 >= len(layout.RowBounds) {
			continue
		}
		centerY := autismDevTrainingEffectPDFY((layout.RowBounds[itemIndex]+layout.RowBounds[itemIndex+1])/2, layout.Height)
		r.drawRowChecks(layout, centerY, row)
	}
	return nil
}

func (r *autismDevTrainingEffectPDFRenderer) drawTemplateBackground(pageNo int) error {
	path := fmt.Sprintf("%s/page_%02d.png", autismDevTrainingEffectTemplateDir, pageNo)
	raw, err := autismDevTrainingEffectTemplateImages.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load AutismDev training template %s: %w", path, err)
	}
	holder, err := gopdf.ImageHolderByBytes(raw)
	if err != nil {
		return fmt.Errorf("decode AutismDev training template %s: %w", path, err)
	}
	if err := r.pdf.ImageByHolder(holder, 0, 0, &gopdf.Rect{W: autismDevProfilePDFPageWidth, H: autismDevProfilePDFPageHeight}); err != nil {
		return fmt.Errorf("draw AutismDev training template %s: %w", path, err)
	}
	return nil
}

func (r *autismDevTrainingEffectPDFRenderer) drawRowChecks(layout autismDevTrainingEffectTemplateLayout, centerY float64, row autismDevTrainingEffectRow) {
	r.drawScoreAt(autismDevTrainingEffectColumnCenter(layout, 2), centerY, row.FirstScore)
	r.drawScoreAt(autismDevTrainingEffectColumnCenter(layout, 3), centerY, row.SecondScore)
	for index, effect := range []string{"significant", "effective", "none"} {
		if row.FirstEffect == effect {
			r.drawEffectAt(autismDevTrainingEffectColumnCenter(layout, 4+index), centerY)
		}
	}
	for index, effect := range []string{"significant", "effective", "none"} {
		if row.SecondEffectReady && row.SecondEffect == effect {
			r.drawEffectAt(autismDevTrainingEffectColumnCenter(layout, 7+index), centerY)
		}
	}
}

func (r *autismDevTrainingEffectPDFRenderer) drawScoreAt(centerX, centerY float64, score string) {
	if strings.TrimSpace(score) == "" || centerX <= 0 || centerY <= 0 {
		return
	}
	r.cellText(
		centerX-autismDevTrainingEffectCheckCellWidth/2,
		centerY-autismDevTrainingEffectCheckCellHeight/2-1,
		autismDevTrainingEffectCheckCellWidth,
		autismDevTrainingEffectScoreFontSize,
		score,
		gopdf.Center|gopdf.Middle,
		0,
		0,
		0,
		true,
	)
}

func (r *autismDevTrainingEffectPDFRenderer) drawEffectAt(centerX, centerY float64) {
	if centerX <= 0 || centerY <= 0 {
		return
	}
	r.pdf.SetLineWidth(1.8)
	r.pdf.SetStrokeColor(0, 0, 0)
	r.pdf.Line(centerX-5.6, centerY+0.8, centerX-1.6, centerY+4.8)
	r.pdf.Line(centerX-1.6, centerY+4.8, centerX+6.4, centerY-5.0)
}

func (r *autismDevTrainingEffectPDFRenderer) cellText(x, y, width, size float64, text string, align int, red, green, blue uint8, bold bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	offsets := []float64{0}
	if bold {
		offsets = []float64{-0.16, 0.16}
	}
	_ = r.pdf.SetFont(autismDevProfilePDFFontFamily, "", size)
	r.pdf.SetTextColor(red, green, blue)
	for _, offset := range offsets {
		r.pdf.SetXY(x+offset, y)
		_ = r.pdf.CellWithOption(&gopdf.Rect{W: width, H: size * 1.65}, text, gopdf.CellOption{Align: align})
	}
}

func autismDevTrainingEffectTemplatePagesForDomain(domainCode string) []autismDevTrainingEffectTemplatePage {
	pages := autismDevTrainingEffectTemplatePages()
	out := make([]autismDevTrainingEffectTemplatePage, 0, 3)
	for _, page := range pages {
		if page.DomainCode == domainCode {
			out = append(out, page)
		}
	}
	return out
}

func autismDevTrainingEffectTemplatePages() []autismDevTrainingEffectTemplatePage {
	return []autismDevTrainingEffectTemplatePage{
		{PageNo: 43, DomainCode: autismdevscore.DomainSensory, FirstItemNo: 1, LastItemNo: 19},
		{PageNo: 44, DomainCode: autismdevscore.DomainSensory, FirstItemNo: 20, LastItemNo: 36},
		{PageNo: 45, DomainCode: autismdevscore.DomainSensory, FirstItemNo: 37, LastItemNo: 55},
		{PageNo: 46, DomainCode: autismdevscore.DomainGrossMotor, FirstItemNo: 1, LastItemNo: 27},
		{PageNo: 47, DomainCode: autismdevscore.DomainGrossMotor, FirstItemNo: 28, LastItemNo: 53},
		{PageNo: 48, DomainCode: autismdevscore.DomainGrossMotor, FirstItemNo: 54, LastItemNo: 72},
		{PageNo: 49, DomainCode: autismdevscore.DomainFineMotor, FirstItemNo: 1, LastItemNo: 21},
		{PageNo: 50, DomainCode: autismdevscore.DomainFineMotor, FirstItemNo: 22, LastItemNo: 43},
		{PageNo: 51, DomainCode: autismdevscore.DomainFineMotor, FirstItemNo: 44, LastItemNo: 66},
		{PageNo: 52, DomainCode: autismdevscore.DomainLanguageComm, FirstItemNo: 1, LastItemNo: 25},
		{PageNo: 53, DomainCode: autismdevscore.DomainLanguageComm, FirstItemNo: 26, LastItemNo: 54},
		{PageNo: 54, DomainCode: autismdevscore.DomainLanguageComm, FirstItemNo: 55, LastItemNo: 79},
		{PageNo: 55, DomainCode: autismdevscore.DomainCognition, FirstItemNo: 1, LastItemNo: 26},
		{PageNo: 56, DomainCode: autismdevscore.DomainCognition, FirstItemNo: 27, LastItemNo: 55},
		{PageNo: 57, DomainCode: autismdevscore.DomainSocial, FirstItemNo: 1, LastItemNo: 16},
		{PageNo: 58, DomainCode: autismdevscore.DomainSocial, FirstItemNo: 17, LastItemNo: 35},
		{PageNo: 59, DomainCode: autismdevscore.DomainSocial, FirstItemNo: 36, LastItemNo: 47},
		{PageNo: 60, DomainCode: autismdevscore.DomainDailyLiving, FirstItemNo: 1, LastItemNo: 22},
		{PageNo: 61, DomainCode: autismdevscore.DomainDailyLiving, FirstItemNo: 23, LastItemNo: 46},
		{PageNo: 62, DomainCode: autismdevscore.DomainDailyLiving, FirstItemNo: 47, LastItemNo: 67},
		{PageNo: 63, DomainCode: autismdevscore.DomainEmotionBehavior, FirstItemNo: 1, LastItemNo: 27},
		{PageNo: 64, DomainCode: autismdevscore.DomainEmotionBehavior, FirstItemNo: 28, LastItemNo: 52},
	}
}

func autismDevTrainingEffectTemplateLayoutForPage(templatePage autismDevTrainingEffectTemplatePage) (autismDevTrainingEffectTemplateLayout, error) {
	autismDevTrainingEffectTemplateLayoutMu.RLock()
	layout, ok := autismDevTrainingEffectTemplateLayoutCache[templatePage.PageNo]
	autismDevTrainingEffectTemplateLayoutMu.RUnlock()
	if ok {
		return layout, nil
	}

	path := fmt.Sprintf("%s/page_%02d.png", autismDevTrainingEffectTemplateDir, templatePage.PageNo)
	file, err := autismDevTrainingEffectTemplateImages.Open(path)
	if err != nil {
		return autismDevTrainingEffectTemplateLayout{}, fmt.Errorf("open AutismDev training template %s: %w", path, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return autismDevTrainingEffectTemplateLayout{}, fmt.Errorf("decode AutismDev training template layout %s: %w", path, err)
	}
	bounds := img.Bounds()
	dark := autismDevTrainingEffectDarkPixelFunc(img)
	columnXs := autismDevTrainingEffectDetectColumnLines(bounds, dark)
	if len(columnXs) < 11 {
		return autismDevTrainingEffectTemplateLayout{}, fmt.Errorf("detect AutismDev training template columns page %d: got %d", templatePage.PageNo, len(columnXs))
	}
	if len(columnXs) > 11 {
		columnXs = autismDevTrainingEffectNormalizeColumnLines(columnXs)
	}
	rowLines := autismDevTrainingEffectDetectHorizontalLines(bounds, dark, int(columnXs[0]), int(columnXs[len(columnXs)-1]))
	itemCount := templatePage.LastItemNo - templatePage.FirstItemNo + 1
	rowStart := len(rowLines) - (itemCount + 1) - templatePage.TrailingHeaders
	if rowStart < 0 || rowStart+itemCount+1 > len(rowLines) {
		return autismDevTrainingEffectTemplateLayout{}, fmt.Errorf("detect AutismDev training template rows page %d: got %d for %d items", templatePage.PageNo, len(rowLines), itemCount)
	}
	rowBounds := append([]float64(nil), rowLines[rowStart:rowStart+itemCount+1]...)
	layout = autismDevTrainingEffectTemplateLayout{
		PageNo:       templatePage.PageNo,
		Width:        bounds.Dx(),
		Height:       bounds.Dy(),
		ColumnXs:     columnXs,
		RowBounds:    rowBounds,
		FirstItemNo:  templatePage.FirstItemNo,
		LastItemNo:   templatePage.LastItemNo,
		Continuation: templatePage.Continuation,
	}

	autismDevTrainingEffectTemplateLayoutMu.Lock()
	autismDevTrainingEffectTemplateLayoutCache[templatePage.PageNo] = layout
	autismDevTrainingEffectTemplateLayoutMu.Unlock()
	return layout, nil
}

func autismDevTrainingEffectDarkPixelFunc(img image.Image) func(int, int) bool {
	return func(x, y int) bool {
		r, g, b, a := img.At(x, y).RGBA()
		if a < 0x8000 {
			return false
		}
		luma := (299*r + 587*g + 114*b) / 1000
		return luma < 0x7000
	}
}

func autismDevTrainingEffectDetectColumnLines(bounds image.Rectangle, dark func(int, int) bool) []float64 {
	minY := bounds.Min.Y + 80
	maxY := bounds.Max.Y - 80
	candidates := make([]int, 0, 16)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		if autismDevTrainingEffectLongestVerticalRun(x, minY, maxY, dark) > 300 {
			candidates = append(candidates, x)
		}
	}
	return autismDevTrainingEffectClusterCenters(candidates, 3)
}

func autismDevTrainingEffectNormalizeColumnLines(xs []float64) []float64 {
	if len(xs) <= 11 {
		return xs
	}
	type window struct {
		start int
		score float64
	}
	best := window{start: 0, score: -1}
	for start := 0; start+11 <= len(xs); start++ {
		candidate := xs[start : start+11]
		score := candidate[10] - candidate[0]
		for index := 1; index < len(candidate); index++ {
			gap := candidate[index] - candidate[index-1]
			if gap < 38 || gap > 230 {
				score -= 500
			}
		}
		if score > best.score {
			best = window{start: start, score: score}
		}
	}
	return append([]float64(nil), xs[best.start:best.start+11]...)
}

func autismDevTrainingEffectDetectHorizontalLines(bounds image.Rectangle, dark func(int, int) bool, minX, maxX int) []float64 {
	width := maxX - minX + 1
	if width <= 0 {
		return nil
	}
	candidates := make([]int, 0, 64)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		count := 0
		for x := minX; x <= maxX; x++ {
			if dark(x, y) {
				count++
			}
		}
		if float64(count) > float64(width)*0.55 {
			candidates = append(candidates, y)
		}
	}
	return autismDevTrainingEffectClusterCenters(candidates, 3)
}

func autismDevTrainingEffectLongestVerticalRun(x, minY, maxY int, dark func(int, int) bool) int {
	best := 0
	current := 0
	for y := minY; y < maxY; y++ {
		if dark(x, y) {
			current++
			if current > best {
				best = current
			}
			continue
		}
		current = 0
	}
	return best
}

func autismDevTrainingEffectClusterCenters(values []int, gap int) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := make([]float64, 0, len(values))
	start := values[0]
	prev := values[0]
	for _, value := range values[1:] {
		if value-prev <= gap {
			prev = value
			continue
		}
		out = append(out, float64(start+prev)/2)
		start = value
		prev = value
	}
	out = append(out, float64(start+prev)/2)
	return out
}

func autismDevTrainingEffectColumnCenter(layout autismDevTrainingEffectTemplateLayout, columnIndex int) float64 {
	if columnIndex < 0 || columnIndex+1 >= len(layout.ColumnXs) {
		return 0
	}
	return autismDevTrainingEffectPDFX((layout.ColumnXs[columnIndex]+layout.ColumnXs[columnIndex+1])/2, layout.Width)
}

func autismDevTrainingEffectPDFX(imageX float64, imageWidth int) float64 {
	if imageWidth <= 0 {
		return 0
	}
	return imageX * autismDevProfilePDFPageWidth / float64(imageWidth)
}

func autismDevTrainingEffectPDFY(imageY float64, imageHeight int) float64 {
	if imageHeight <= 0 {
		return 0
	}
	return imageY * autismDevProfilePDFPageHeight / float64(imageHeight)
}

func autismDevTrainingEffectRecords(records []model.AssessmentRecordDetailVO, current model.AssessmentRecordDetailVO) ([]autismDevTrainingEffectRecord, error) {
	byID := make(map[int64]model.AssessmentRecordDetailVO, len(records)+1)
	for _, record := range records {
		if record.ID > 0 {
			byID[record.ID] = record
		}
	}
	if current.ID > 0 {
		byID[current.ID] = current
	}
	ordered := make([]model.AssessmentRecordDetailVO, 0, len(byID))
	for _, record := range byID {
		ordered = append(ordered, record)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		leftDate := ordered[i].AssessmentDate
		rightDate := ordered[j].AssessmentDate
		if leftDate != nil && rightDate != nil && !leftDate.Equal(*rightDate) {
			return leftDate.Before(*rightDate)
		}
		if leftDate != nil && rightDate == nil {
			return true
		}
		if leftDate == nil && rightDate != nil {
			return false
		}
		return ordered[i].ID < ordered[j].ID
	})
	if len(ordered) > 3 {
		ordered = ordered[:3]
	}
	out := make([]autismDevTrainingEffectRecord, 0, len(ordered))
	for index, record := range ordered {
		scores, err := decodeSavedAutismDevInputScores(record.InputJSON)
		if err != nil {
			return nil, err
		}
		record.AssessmentSequence = index + 1
		out = append(out, autismDevTrainingEffectRecord{Record: record, Scores: scores})
	}
	return out, nil
}

func autismDevTrainingEffectRowsByDomain(items []autismdevscore.ItemDefinition, domains []string, records []autismDevTrainingEffectRecord) map[string][]autismDevTrainingEffectRow {
	itemsByDomain := autismDevItemsByDomain(items)
	out := make(map[string][]autismDevTrainingEffectRow, len(domains))
	for _, domainCode := range domains {
		for _, item := range itemsByDomain[domainCode] {
			row := autismDevTrainingEffectRow{Item: item}
			if len(records) >= 1 {
				score := records[0].Scores[item.ItemNo]
				row.FirstScore = autismDevTrainingScoreText(score)
			}
			if len(records) >= 2 {
				score := records[1].Scores[item.ItemNo]
				row.SecondScore = autismDevTrainingScoreText(score)
				row.FirstEffect = autismDevTrainingEffectForScores(item.ScoreType, records[0].Scores[item.ItemNo], score)
			}
			if len(records) >= 3 {
				if effect := autismDevTrainingEffectForScores(item.ScoreType, records[1].Scores[item.ItemNo], records[2].Scores[item.ItemNo]); effect != "" {
					row.SecondEffectReady = true
					row.SecondEffect = effect
				}
			}
			out[domainCode] = append(out[domainCode], row)
		}
	}
	return out
}

func autismDevTrainingScoreText(score string) string {
	return normalizeAutismDevScore(score)
}

func autismDevTrainingProjectChecked(scoreType, score string) bool {
	normalized := normalizeAutismDevScore(score)
	return normalized != "" && normalized != autismdevscore.ScoreX
}

func autismDevTrainingEffectForScores(scoreType, before, after string) string {
	before = normalizeAutismDevScore(before)
	after = normalizeAutismDevScore(after)
	if before == "" || after == "" || before == autismdevscore.ScoreX {
		return ""
	}
	beforeRank, okBefore := autismDevTrainingScoreRank(scoreType, before)
	afterRank, okAfter := autismDevTrainingScoreRank(scoreType, after)
	if !okBefore || !okAfter {
		return ""
	}
	if afterRank <= beforeRank {
		return "none"
	}
	if afterRank-beforeRank >= 2 {
		return "significant"
	}
	return "effective"
}

func autismDevTrainingScoreRank(scoreType, score string) (int, bool) {
	switch strings.ToUpper(strings.TrimSpace(scoreType)) {
	case autismdevscore.ScoreTypeAMS:
		switch normalizeAutismDevScore(score) {
		case autismdevscore.ScoreX:
			return 0, true
		case autismdevscore.ScoreS:
			return 1, true
		case autismdevscore.ScoreM:
			return 2, true
		case autismdevscore.ScoreA:
			return 3, true
		default:
			return 0, false
		}
	default:
		switch normalizeAutismDevScore(score) {
		case autismdevscore.ScoreX:
			return 0, true
		case autismdevscore.ScoreF:
			return 1, true
		case autismdevscore.ScoreE:
			return 2, true
		case autismdevscore.ScoreP:
			return 3, true
		default:
			return 0, false
		}
	}
}

func autismDevTrainingCurrentRecordDomains(record model.AssessmentRecordDetailVO, data autismDevStaticData) []string {
	known := autismDevKnownDomainSet(data.domains)
	if scoped := decodeAutismDevTrainingScopeDomains(record.InputJSON, known); len(scoped) > 0 {
		return scoped
	}
	itemByNo := make(map[int]autismdevscore.ItemDefinition, len(data.items))
	for _, item := range data.items {
		itemByNo[item.ItemNo] = item
	}
	scores, _ := decodeSavedAutismDevInputScores(record.InputJSON)
	seen := make(map[string]bool, len(known))
	for itemNo := range scores {
		item, ok := itemByNo[itemNo]
		if !ok {
			continue
		}
		code := strings.TrimSpace(item.DomainCode)
		if code != "" {
			seen[code] = true
		}
	}
	ordered := make([]string, 0, len(seen))
	for _, domainCode := range autismDevDomainCodesFromDefinitions(data.domains) {
		if seen[domainCode] {
			ordered = append(ordered, domainCode)
		}
	}
	return ordered
}

func decodeAutismDevTrainingScopeDomains(raw json.RawMessage, known map[string]bool) []string {
	if len(raw) == 0 {
		return nil
	}
	var scope autismDevTrainingInputScope
	if err := json.Unmarshal(raw, &scope); err != nil {
		return nil
	}
	mode := strings.ToLower(strings.TrimSpace(nonEmptyString(scope.ScopeMode, scope.AssessmentScopeMode)))
	codes := scope.ScopeDomainCodes
	if len(codes) == 0 {
		codes = scope.SelectedDomainCodes
	}
	if mode != "custom" && len(codes) == 0 {
		return nil
	}
	out := make([]string, 0, len(codes))
	seen := make(map[string]bool, len(codes))
	for _, rawCode := range codes {
		code := strings.TrimSpace(rawCode)
		if code == "" || !known[code] || seen[code] {
			continue
		}
		seen[code] = true
		out = append(out, code)
	}
	return out
}

func autismDevKnownDomainSet(domains []autismDevDomainDefinition) map[string]bool {
	out := make(map[string]bool, len(domains))
	for _, domain := range domains {
		code := strings.TrimSpace(domain.ScaleCode)
		if code != "" {
			out[code] = true
		}
	}
	return out
}

func autismDevDomainCodesFromDefinitions(domains []autismDevDomainDefinition) []string {
	out := make([]string, 0, len(domains))
	for _, domain := range domains {
		code := strings.TrimSpace(domain.ScaleCode)
		if code != "" {
			out = append(out, code)
		}
	}
	return out
}

func autismDevTrainingDomainName(domainCode string, data autismDevStaticData) string {
	for _, domain := range data.domains {
		if strings.TrimSpace(domain.ScaleCode) == domainCode {
			return autismDevDomainDisplayName(domain.ScaleName, domainCode)
		}
	}
	for _, item := range data.items {
		if strings.TrimSpace(item.DomainCode) == domainCode {
			return autismDevDomainDisplayName(item.DomainName, domainCode)
		}
	}
	return domainCode
}

func autismDevTrainingManualSection(domainCode string) string {
	for index, code := range autismdevscore.DomainOrder {
		if code == domainCode {
			return fmt.Sprintf("4.%d", index+1)
		}
	}
	return "4"
}
