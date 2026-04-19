package service

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

//go:embed templates/rehab-record-template.docx
var rehabRecordTemplateDocx []byte

const (
	classCommentWordExportDocumentPath     = "word/document.xml"
	classCommentWordExportDocumentRelsPath = "word/_rels/document.xml.rels"
	classCommentWordExportContentTypesPath = "[Content_Types].xml"

	classCommentWordExportImageRelationshipType = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
	classCommentWordExportSignatureTabPos       = 7000
	classCommentWordExportSignatureHeightEMU    = 203200
	classCommentWordExportSignatureMaxWidthEMU  = 1270000
	classCommentWordExportParagraphBefore       = 60
	classCommentWordExportParagraphAfter        = 60
)

var wordTCVerticalAlignRegexp = regexp.MustCompile(`<w:vAlign\b[^>]*w:val="[^"]*"[^>]*/>`)
var wordTCWidthRegexp = regexp.MustCompile(`<w:tcW\b[^>]*w:w="[^"]+"[^>]*/>`)
var wordGridColWidthRegexp = regexp.MustCompile(`<w:gridCol\b[^>]*w:w="[^"]+"[^>]*/>`)
var wordTCNoWrapRegexp = regexp.MustCompile(`<w:noWrap\b[^>]*/>`)
var wordTRHeightRegexp = regexp.MustCompile(`<w:trHeight\b[^>]*w:val="[^"]+"[^>]*/>`)

var classCommentWordExportFirstRowWidths = []int{866, 1396, 726, 620, 1214, 1435, 484, 1775}

type docxRelationships struct {
	XMLName       xml.Name           `xml:"Relationships"`
	XMLNS         string             `xml:"xmlns,attr"`
	Relationships []docxRelationship `xml:"Relationship"`
}

type docxRelationship struct {
	ID     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
}

type docxContentTypes struct {
	XMLName   xml.Name                  `xml:"Types"`
	XMLNS     string                    `xml:"xmlns,attr"`
	Defaults  []docxContentTypeDefault  `xml:"Default"`
	Overrides []docxContentTypeOverride `xml:"Override"`
}

type docxContentTypeDefault struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

type docxContentTypeOverride struct {
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}

type classCommentWordDocxMedia struct {
	RelationshipID string
	Target         string
	ZipPath        string
	ContentType    string
	FileName       string
	Data           []byte
	WidthEMU       int64
	HeightEMU      int64
	DocPrID        int
}

type classCommentWordDocxBuildState struct {
	nextRelationshipIndex int
	nextDocPrID           int
	media                 []classCommentWordDocxMedia
}

func validateClassCommentWordExportQuery(query model.ClassCommentQueryModel) error {
	if strings.TrimSpace(query.StudentID) == "" {
		return errors.New("导出前请先筛选学员")
	}

	startText := strings.TrimSpace(query.TeachingStartTime)
	endText := strings.TrimSpace(query.TeachingEndTime)
	if startText == "" || endText == "" {
		return errors.New("导出前请先筛选上课日期")
	}

	start, err := time.ParseInLocation("2006-01-02", startText, time.Local)
	if err != nil {
		return errors.New("上课开始日期格式不正确")
	}
	end, err := time.ParseInLocation("2006-01-02", endText, time.Local)
	if err != nil {
		return errors.New("上课结束日期格式不正确")
	}
	if end.Before(start) {
		return errors.New("上课日期筛选范围不正确")
	}
	if end.After(start.AddDate(0, 1, 0)) {
		return errors.New("导出时间范围最大支持一个月")
	}
	return nil
}

func buildClassCommentWordDocx(items []rehabRecordWordExportView) ([]byte, error) {
	if len(items) == 0 {
		return nil, errors.New("暂无可导出的康复记录")
	}

	entries, err := readDocxZipEntries(rehabRecordTemplateDocx)
	if err != nil {
		return nil, err
	}

	documentXML := string(entries[classCommentWordExportDocumentPath])
	titleXML, tableXML, spacerXML, sectPrXML, err := extractTemplateDocumentParts(documentXML)
	if err != nil {
		return nil, err
	}

	rels, err := parseDocxRelationships(entries[classCommentWordExportDocumentRelsPath])
	if err != nil {
		return nil, err
	}
	contentTypes, err := parseDocxContentTypes(entries[classCommentWordExportContentTypesPath])
	if err != nil {
		return nil, err
	}

	state := classCommentWordDocxBuildState{
		nextRelationshipIndex: nextDocxRelationshipIndex(rels.Relationships),
		nextDocPrID:           1,
		media:                 make([]classCommentWordDocxMedia, 0),
	}

	bodyChildren := make([]string, 0, len(items)*4+1)
	for index, item := range items {
		if index > 0 {
			bodyChildren = append(bodyChildren, buildWordPageBreakParagraph())
		}

		recordTableXML, err := buildClassCommentRecordTableXML(tableXML, item, &state)
		if err != nil {
			return nil, err
		}

		bodyChildren = append(bodyChildren, titleXML, recordTableXML, spacerXML)
	}
	bodyChildren = append(bodyChildren, sectPrXML)

	updatedDocumentXML, err := replaceXMLContainerInner(documentXML, "w:body", strings.Join(bodyChildren, ""))
	if err != nil {
		return nil, err
	}
	entries[classCommentWordExportDocumentPath] = []byte(updatedDocumentXML)

	for _, media := range state.media {
		rels.Relationships = append(rels.Relationships, docxRelationship{
			ID:     media.RelationshipID,
			Type:   classCommentWordExportImageRelationshipType,
			Target: media.Target,
		})
		contentTypes = ensureDocxContentTypeDefault(contentTypes, fileExtensionWithoutDot(media.FileName), media.ContentType)
		entries[media.ZipPath] = media.Data
	}

	relsXML, err := marshalXMLDocument(rels)
	if err != nil {
		return nil, err
	}
	contentTypesXML, err := marshalXMLDocument(contentTypes)
	if err != nil {
		return nil, err
	}

	entries[classCommentWordExportDocumentRelsPath] = relsXML
	entries[classCommentWordExportContentTypesPath] = contentTypesXML

	return writeDocxZipEntries(entries)
}

func buildClassCommentRecordTableXML(tableXML string, item rehabRecordWordExportView, state *classCommentWordDocxBuildState) (string, error) {
	_, tableInner, _, err := splitXMLContainer(tableXML, "w:tbl")
	if err != nil {
		return "", err
	}
	children, err := splitTopLevelXML(tableInner)
	if err != nil {
		return "", err
	}
	if len(children) < 12 {
		return "", errors.New("康复记录模板结构不正确")
	}

	for index, width := range classCommentWordExportFirstRowWidths {
		children[1] = setWordTableGridColumnWidth(children[1], index, width)
	}
	children[2] = setRowCellWidth(children[2], 1, classCommentWordExportFirstRowWidths[1])
	children[2] = setRowCellWidth(children[2], 3, classCommentWordExportFirstRowWidths[3])
	children[2] = setRowCellWidth(children[2], 4, classCommentWordExportFirstRowWidths[4])
	children[2] = setRowCellWidth(children[2], 5, classCommentWordExportFirstRowWidths[5])
	children[2] = setRowCellWidth(children[2], 7, classCommentWordExportFirstRowWidths[7])
	children[2] = setRowCellNoWrap(children[2], 5, true)
	children[3] = setRowCellWidth(children[3], 1, sumWordTableWidths(classCommentWordExportFirstRowWidths, 1, 3))
	children[3] = setRowCellWidth(children[3], 2, classCommentWordExportFirstRowWidths[4])
	children[3] = setRowCellWidth(children[3], 3, sumWordTableWidths(classCommentWordExportFirstRowWidths, 5, 7))

	row0 := replaceCellContent(children[2], 1, buildWordParagraphs(strings.TrimSpace(item.StudentName), "left", "", 240), "")
	row0 = replaceCellContent(row0, 3, buildWordParagraphs(strings.TrimSpace(item.Gender), "center", "", 240), "")
	row0 = replaceCellContent(row0, 5, buildWordParagraphs(strings.TrimSpace(item.BirthDate), "left", "", 240), "")
	row0 = replaceCellContent(row0, 7, buildWordParagraphs(strings.TrimSpace(item.ClassName), "left", "", 240), "")

	row1 := replaceCellContent(children[3], 1, buildWordParagraphs(strings.TrimSpace(item.TeacherName), "left", "", 240), "")
	row1 = replaceCellContent(row1, 3, buildWordParagraphs(strings.TrimSpace(item.TrainingDate), "left", "", 240), "")

	row2 := replaceCellContent(children[4], 1, buildWordParagraphs(strings.TrimSpace(item.TrainingTarget), "left", "", 240), "top")
	row3 := children[5]

	trainingRows := buildClassCommentTrainingRows(children[6], children[7], item.TrainingItems)
	row6 := replaceCellContent(children[8], 1, buildWordParagraphs(strings.TrimSpace(item.Performance), "left", "", 240), "center")
	row7 := setWordTableRowHeight(children[9], 1120)
	row7 = replaceCellContent(row7, 1, buildWordParagraphs(strings.TrimSpace(item.Suggestion), "left", "", 240), "center")
	row8 := setWordTableRowHeight(children[10], 1120)
	row8 = replaceCellContent(row8, 1, buildWordParagraphs(strings.TrimSpace(item.ParentFeedback), "left", "", 240), "top")
	row9, err := buildClassCommentSignatureRow(children[11], item, state)
	if err != nil {
		return "", err
	}

	finalChildren := make([]string, 0, 2+len(trainingRows)+7)
	finalChildren = append(finalChildren, children[0], children[1], row0, row1, row2, row3)
	finalChildren = append(finalChildren, trainingRows...)
	finalChildren = append(finalChildren, row6, row7, row8, row9)

	return replaceXMLContainerInner(tableXML, "w:tbl", strings.Join(finalChildren, ""))
}

func buildClassCommentTrainingRows(firstTemplateRow, repeatTemplateRow string, items []model.RehabRecordTrainingItem) []string {
	if len(items) == 0 {
		items = []model.RehabRecordTrainingItem{{}}
	}

	rows := make([]string, 0, len(items))
	for index, item := range items {
		templateRow := repeatTemplateRow
		if index == 0 {
			templateRow = firstTemplateRow
		}
		rowXML := replaceCellContent(templateRow, 0, buildWordParagraphs(formatTrainingItemTitle(item.Title), "center", "", 240), "")
		rowXML = replaceCellContent(rowXML, 1, buildWordParagraphs(strings.TrimSpace(item.Content), "left", "", 240), "top")
		rows = append(rows, rowXML)
	}
	return rows
}

func buildClassCommentSignatureRow(rowXML string, item rehabRecordWordExportView, state *classCommentWordDocxBuildState) (string, error) {
	signatureParagraph, err := buildClassCommentSignatureParagraph(item, state)
	if err != nil {
		return "", err
	}

	return replaceCellContent(rowXML, 1, []string{signatureParagraph}, ""), nil
}

func buildClassCommentSignatureParagraph(item rehabRecordWordExportView, state *classCommentWordDocxBuildState) (string, error) {
	var builder strings.Builder
	builder.WriteString(`<w:p><w:pPr><w:tabs><w:tab w:pos="`)
	builder.WriteString(strconv.Itoa(classCommentWordExportSignatureTabPos))
	builder.WriteString(`" w:val="right"/></w:tabs><w:jc w:val="both"/>`)
	builder.WriteString(defaultWordParagraphRunPropsXML())
	builder.WriteString(`</w:pPr>`)
	builder.WriteString(buildWordTextRunXML("家长签名："))

	signatureXML, err := buildClassCommentSignatureContentXML(item, state)
	if err != nil {
		return "", err
	}
	builder.WriteString(signatureXML)
	builder.WriteString(`<w:r><w:tab/></w:r>`)
	builder.WriteString(buildWordTextRunXML(formatDocxFeedbackDateText(item.FeedbackDate)))
	builder.WriteString(`</w:p>`)
	return builder.String(), nil
}

func buildClassCommentSignatureContentXML(item rehabRecordWordExportView, state *classCommentWordDocxBuildState) (string, error) {
	if asset, ok := tryAppendClassCommentSignatureImage(item.ParentSignatureImage, state); ok {
		return buildWordDrawingRunXML(asset), nil
	}

	text := strings.TrimSpace(item.ParentSignatureText)
	if text == "" {
		return "", nil
	}
	return buildWordTextRunXML(text), nil
}

func tryAppendClassCommentSignatureImage(raw string, state *classCommentWordDocxBuildState) (classCommentWordDocxMedia, bool) {
	source := strings.TrimSpace(raw)
	if source == "" {
		return classCommentWordDocxMedia{}, false
	}

	asset, err := loadClassCommentSignatureImage(source, state)
	if err != nil {
		return classCommentWordDocxMedia{}, false
	}
	return asset, true
}

func loadClassCommentSignatureImage(source string, state *classCommentWordDocxBuildState) (classCommentWordDocxMedia, error) {
	data, contentType, fileName, err := fetchClassCommentExportImage(source)
	if err != nil {
		return classCommentWordDocxMedia{}, err
	}
	if len(data) == 0 {
		return classCommentWordDocxMedia{}, errors.New("empty signature image")
	}

	widthEMU, heightEMU, err := calcClassCommentSignatureImageEMU(data)
	if err != nil {
		return classCommentWordDocxMedia{}, err
	}

	extension := fileExtensionWithoutDot(fileName)
	if extension == "" {
		extension = extensionByContentType(contentType)
	}
	if extension == "" {
		return classCommentWordDocxMedia{}, errors.New("unsupported signature image type")
	}

	state.nextRelationshipIndex++
	state.nextDocPrID++
	fileName = fmt.Sprintf("rehab-signature-%d.%s", len(state.media)+1, extension)
	asset := classCommentWordDocxMedia{
		RelationshipID: fmt.Sprintf("rId%d", state.nextRelationshipIndex),
		Target:         "media/" + fileName,
		ZipPath:        "word/media/" + fileName,
		ContentType:    normalizeContentTypeByExtension(contentType, extension),
		FileName:       fileName,
		Data:           data,
		WidthEMU:       widthEMU,
		HeightEMU:      heightEMU,
		DocPrID:        state.nextDocPrID,
	}
	state.media = append(state.media, asset)
	return asset, nil
}

func fetchClassCommentExportImage(source string) ([]byte, string, string, error) {
	if strings.HasPrefix(strings.ToLower(source), "data:image/") {
		return decodeClassCommentExportDataImage(source)
	}

	request, err := http.NewRequest(http.MethodGet, source, nil)
	if err != nil {
		return nil, "", "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, "", "", err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", "", fmt.Errorf("fetch image failed: %s", response.Status)
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, 10<<20))
	if err != nil {
		return nil, "", "", err
	}

	contentType := strings.TrimSpace(response.Header.Get("Content-Type"))
	if mediaType, _, parseErr := mime.ParseMediaType(contentType); parseErr == nil {
		contentType = mediaType
	}

	fileName := filepath.Base(strings.Split(source, "?")[0])
	return data, contentType, fileName, nil
}

func decodeClassCommentExportDataImage(source string) ([]byte, string, string, error) {
	prefix, encoded, found := strings.Cut(source, ",")
	if !found {
		return nil, "", "", errors.New("invalid data image")
	}

	contentType := strings.TrimSpace(strings.TrimPrefix(prefix, "data:"))
	if index := strings.Index(contentType, ";"); index >= 0 {
		contentType = contentType[:index]
	}
	if contentType == "" {
		return nil, "", "", errors.New("invalid data image content type")
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", "", err
	}

	extension := extensionByContentType(contentType)
	fileName := "signature"
	if extension != "" {
		fileName += "." + extension
	}
	return data, contentType, fileName, nil
}

func calcClassCommentSignatureImageEMU(data []byte) (int64, int64, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}
	if config.Width <= 0 || config.Height <= 0 {
		return 0, 0, errors.New("invalid signature image size")
	}

	heightEMU := int64(classCommentWordExportSignatureHeightEMU)
	widthEMU := int64(config.Width) * heightEMU / int64(config.Height)
	if widthEMU > classCommentWordExportSignatureMaxWidthEMU {
		widthEMU = classCommentWordExportSignatureMaxWidthEMU
	}
	if widthEMU <= 0 {
		widthEMU = heightEMU
	}
	return widthEMU, heightEMU, nil
}

func buildWordDrawingRunXML(asset classCommentWordDocxMedia) string {
	var builder strings.Builder
	builder.WriteString(`<w:r><w:drawing><wp:inline xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">`)
	builder.WriteString(`<wp:extent cx="`)
	builder.WriteString(strconv.FormatInt(asset.WidthEMU, 10))
	builder.WriteString(`" cy="`)
	builder.WriteString(strconv.FormatInt(asset.HeightEMU, 10))
	builder.WriteString(`"/>`)
	builder.WriteString(`<wp:docPr id="`)
	builder.WriteString(strconv.Itoa(asset.DocPrID))
	builder.WriteString(`" name="`)
	builder.WriteString(escapeXMLText(asset.FileName))
	builder.WriteString(`"/>`)
	builder.WriteString(`<wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr>`)
	builder.WriteString(`<a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic>`)
	builder.WriteString(`<pic:nvPicPr><pic:cNvPr id="0" name="`)
	builder.WriteString(escapeXMLText(asset.FileName))
	builder.WriteString(`"/><pic:cNvPicPr/></pic:nvPicPr>`)
	builder.WriteString(`<pic:blipFill><a:blip r:embed="`)
	builder.WriteString(asset.RelationshipID)
	builder.WriteString(`"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>`)
	builder.WriteString(`<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="`)
	builder.WriteString(strconv.FormatInt(asset.WidthEMU, 10))
	builder.WriteString(`" cy="`)
	builder.WriteString(strconv.FormatInt(asset.HeightEMU, 10))
	builder.WriteString(`"/></a:xfrm><a:prstGeom prst="rect"/></pic:spPr>`)
	builder.WriteString(`</pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>`)
	return builder.String()
}

func buildWordPageBreakParagraph() string {
	return `<w:p><w:r><w:br w:type="page"/></w:r></w:p>`
}

func buildWordParagraphs(text, align, tabXML string, lineHeight int) []string {
	return []string{buildWordParagraphXML(text, align, tabXML, lineHeight)}
}

func buildWordParagraphXML(text, align, tabXML string, lineHeight int) string {
	var builder strings.Builder
	builder.WriteString(`<w:p><w:pPr>`)
	if lineHeight > 0 {
		builder.WriteString(`<w:spacing w:before="`)
		builder.WriteString(strconv.Itoa(classCommentWordExportParagraphBefore))
		builder.WriteString(`" w:after="`)
		builder.WriteString(strconv.Itoa(classCommentWordExportParagraphAfter))
		builder.WriteString(`" w:line="`)
		builder.WriteString(strconv.Itoa(lineHeight))
		builder.WriteString(`" w:lineRule="auto"/>`)
	}
	if tabXML != "" {
		builder.WriteString(tabXML)
	}
	if align != "" {
		builder.WriteString(`<w:jc w:val="`)
		builder.WriteString(align)
		builder.WriteString(`"/>`)
	}
	builder.WriteString(defaultWordParagraphRunPropsXML())
	builder.WriteString(`</w:pPr>`)
	if strings.TrimSpace(text) != "" || strings.Contains(text, "\n") {
		builder.WriteString(buildWordTextRunsXML(text))
	}
	builder.WriteString(`</w:p>`)
	return builder.String()
}

func defaultWordParagraphRunPropsXML() string {
	return `<w:rPr><w:rFonts w:hint="default"/><w:vertAlign w:val="baseline"/><w:lang w:val="en-US" w:eastAsia="zh-CN"/></w:rPr>`
}

func defaultWordRunPropsXML() string {
	return `<w:rPr><w:rFonts w:hint="eastAsia"/><w:vertAlign w:val="baseline"/><w:lang w:val="en-US" w:eastAsia="zh-CN"/></w:rPr>`
}

func buildWordTextRunsXML(text string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), "\n")
	var builder strings.Builder
	for index, line := range lines {
		if index > 0 {
			builder.WriteString(`<w:r>`)
			builder.WriteString(defaultWordRunPropsXML())
			builder.WriteString(`<w:br w:type="textWrapping"/></w:r>`)
		}
		builder.WriteString(buildWordTextRunXML(line))
	}
	return builder.String()
}

func buildWordTextRunXML(text string) string {
	var builder strings.Builder
	builder.WriteString(`<w:r>`)
	builder.WriteString(defaultWordRunPropsXML())
	builder.WriteString(`<w:t`)
	if shouldPreserveWordTextSpaces(text) {
		builder.WriteString(` xml:space="preserve"`)
	}
	builder.WriteString(`>`)
	builder.WriteString(escapeXMLText(text))
	builder.WriteString(`</w:t></w:r>`)
	return builder.String()
}

func shouldPreserveWordTextSpaces(text string) bool {
	return strings.HasPrefix(text, " ") || strings.HasSuffix(text, " ") || strings.Contains(text, "  ")
}

func formatDocxFeedbackDateText(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "年      月      日"
	}
	return formatDisplayDateForDocument(text)
}

func formatTrainingItemTitle(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" || strings.Contains(text, "\n") {
		return text
	}

	runes := []rune(text)
	if len(runes) <= 2 {
		return text
	}

	var builder strings.Builder
	for index := 0; index < len(runes); index += 2 {
		if index > 0 {
			builder.WriteByte('\n')
		}
		end := index + 2
		if end > len(runes) {
			end = len(runes)
		}
		builder.WriteString(string(runes[index:end]))
	}
	return builder.String()
}

func replaceRowCellContents(rowXML string, replacements map[int]string) string {
	if len(replacements) == 0 {
		return rowXML
	}

	startTag, innerXML, endTag, err := splitXMLContainer(rowXML, "w:tr")
	if err != nil {
		return rowXML
	}
	children, err := splitTopLevelXML(innerXML)
	if err != nil {
		return rowXML
	}

	cellIndex := 0
	for index, child := range children {
		if !xmlFragmentHasStartTag(child, "w:tc") {
			continue
		}
		if replacement, ok := replacements[cellIndex]; ok {
			children[index] = replacement
		}
		cellIndex++
	}
	return startTag + strings.Join(children, "") + endTag
}

func replaceCellContent(rowXML string, targetCellIndex int, paragraphs []string, verticalAlign string) string {
	startTag, innerXML, endTag, err := splitXMLContainer(rowXML, "w:tr")
	if err != nil {
		return rowXML
	}
	children, err := splitTopLevelXML(innerXML)
	if err != nil {
		return rowXML
	}

	cellIndex := 0
	for index, child := range children {
		if !xmlFragmentHasStartTag(child, "w:tc") {
			continue
		}
		if cellIndex == targetCellIndex {
			children[index] = replaceWordTableCellContent(child, paragraphs, verticalAlign)
			break
		}
		cellIndex++
	}
	return startTag + strings.Join(children, "") + endTag
}

func setRowCellWidth(rowXML string, targetCellIndex int, width int) string {
	startTag, innerXML, endTag, err := splitXMLContainer(rowXML, "w:tr")
	if err != nil {
		return rowXML
	}
	children, err := splitTopLevelXML(innerXML)
	if err != nil {
		return rowXML
	}

	cellIndex := 0
	for index, child := range children {
		if !xmlFragmentHasStartTag(child, "w:tc") {
			continue
		}
		if cellIndex == targetCellIndex {
			children[index] = setWordTableCellWidth(child, width)
			break
		}
		cellIndex++
	}
	return startTag + strings.Join(children, "") + endTag
}

func setRowCellNoWrap(rowXML string, targetCellIndex int, enabled bool) string {
	startTag, innerXML, endTag, err := splitXMLContainer(rowXML, "w:tr")
	if err != nil {
		return rowXML
	}
	children, err := splitTopLevelXML(innerXML)
	if err != nil {
		return rowXML
	}

	cellIndex := 0
	for index, child := range children {
		if !xmlFragmentHasStartTag(child, "w:tc") {
			continue
		}
		if cellIndex == targetCellIndex {
			children[index] = setWordTableCellNoWrap(child, enabled)
			break
		}
		cellIndex++
	}
	return startTag + strings.Join(children, "") + endTag
}

func replaceWordTableCellContent(cellXML string, paragraphs []string, verticalAlign string) string {
	startTag, innerXML, endTag, err := splitXMLContainer(cellXML, "w:tc")
	if err != nil {
		return cellXML
	}
	children, err := splitTopLevelXML(innerXML)
	if err != nil {
		return cellXML
	}

	newChildren := make([]string, 0, len(paragraphs)+1)
	if len(children) > 0 && xmlFragmentHasStartTag(children[0], "w:tcPr") {
		tcPr := children[0]
		if verticalAlign != "" {
			tcPr = setWordTCVerticalAlign(tcPr, verticalAlign)
		}
		newChildren = append(newChildren, tcPr)
	}
	if len(paragraphs) == 0 {
		paragraphs = []string{buildWordParagraphXML("", "left", "", 240)}
	}
	newChildren = append(newChildren, paragraphs...)
	return startTag + strings.Join(newChildren, "") + endTag
}

func setWordTCVerticalAlign(tcPrXML, align string) string {
	if align == "" {
		return tcPrXML
	}
	replacement := `<w:vAlign w:val="` + align + `"/>`
	if wordTCVerticalAlignRegexp.MatchString(tcPrXML) {
		return wordTCVerticalAlignRegexp.ReplaceAllString(tcPrXML, replacement)
	}
	return strings.Replace(tcPrXML, `</w:tcPr>`, replacement+`</w:tcPr>`, 1)
}

func setWordTableCellWidth(cellXML string, width int) string {
	return wordTCWidthRegexp.ReplaceAllString(cellXML, `<w:tcW w:w="`+strconv.Itoa(width)+`" w:type="dxa" />`)
}

func setWordTableCellNoWrap(cellXML string, enabled bool) string {
	startTag, innerXML, endTag, err := splitXMLContainer(cellXML, "w:tc")
	if err != nil {
		return cellXML
	}
	children, err := splitTopLevelXML(innerXML)
	if err != nil {
		return cellXML
	}
	if len(children) == 0 || !xmlFragmentHasStartTag(children[0], "w:tcPr") {
		return cellXML
	}

	tcPr := children[0]
	switch {
	case enabled && wordTCNoWrapRegexp.MatchString(tcPr):
	case enabled:
		tcPr = strings.Replace(tcPr, `</w:tcPr>`, `<w:noWrap/></w:tcPr>`, 1)
	default:
		tcPr = wordTCNoWrapRegexp.ReplaceAllString(tcPr, "")
	}
	children[0] = tcPr
	return startTag + strings.Join(children, "") + endTag
}

func setWordTableGridColumnWidth(gridXML string, targetColumnIndex int, width int) string {
	matches := wordGridColWidthRegexp.FindAllStringIndex(gridXML, -1)
	if targetColumnIndex < 0 || targetColumnIndex >= len(matches) {
		return gridXML
	}

	match := matches[targetColumnIndex]
	return gridXML[:match[0]] + `<w:gridCol w:w="` + strconv.Itoa(width) + `" />` + gridXML[match[1]:]
}

func setWordTableRowHeight(rowXML string, height int) string {
	return wordTRHeightRegexp.ReplaceAllString(rowXML, `<w:trHeight w:val="`+strconv.Itoa(height)+`" w:hRule="atLeast" />`)
}

func sumWordTableWidths(widths []int, start, end int) int {
	total := 0
	if start < 0 {
		start = 0
	}
	if end >= len(widths) {
		end = len(widths) - 1
	}
	for index := start; index <= end; index++ {
		total += widths[index]
	}
	return total
}

func extractTemplateDocumentParts(documentXML string) (string, string, string, string, error) {
	_, bodyInner, _, err := splitXMLContainer(documentXML, "w:body")
	if err != nil {
		return "", "", "", "", err
	}

	children, err := splitTopLevelXML(bodyInner)
	if err != nil {
		return "", "", "", "", err
	}
	if len(children) < 4 {
		return "", "", "", "", errors.New("康复记录模板结构不正确")
	}

	return children[0], children[1], children[2], children[len(children)-1], nil
}

func splitXMLContainer(xmlText, tag string) (string, string, string, error) {
	startIndex := strings.Index(xmlText, "<"+tag)
	if startIndex < 0 {
		return "", "", "", fmt.Errorf("missing <%s>", tag)
	}

	startTagEnd, err := findXMLTagEnd(xmlText, startIndex)
	if err != nil {
		return "", "", "", err
	}
	endTag := "</" + tag + ">"
	endIndex := strings.LastIndex(xmlText, endTag)
	if endIndex < 0 || endIndex < startTagEnd {
		return "", "", "", fmt.Errorf("missing </%s>", tag)
	}

	return xmlText[startIndex:startTagEnd], xmlText[startTagEnd:endIndex], endTag, nil
}

func replaceXMLContainerInner(xmlText, tag, newInner string) (string, error) {
	startTag, _, endTag, err := splitXMLContainer(xmlText, tag)
	if err != nil {
		return "", err
	}
	startIndex := strings.Index(xmlText, startTag)
	if startIndex < 0 {
		return "", fmt.Errorf("missing start tag <%s>", tag)
	}
	endIndex := strings.LastIndex(xmlText, endTag)
	if endIndex < 0 {
		return "", fmt.Errorf("missing end tag </%s>", tag)
	}
	return xmlText[:startIndex] + startTag + newInner + xmlText[endIndex:], nil
}

func splitTopLevelXML(inner string) ([]string, error) {
	result := make([]string, 0)
	depth := 0
	startIndex := -1

	for index := 0; index < len(inner); {
		if depth == 0 {
			for index < len(inner) && isWordXMLWhitespace(inner[index]) {
				index++
			}
			if index >= len(inner) {
				break
			}
			if inner[index] != '<' {
				return nil, errors.New("unexpected text node in template")
			}
			startIndex = index
		}

		if inner[index] != '<' {
			index++
			continue
		}

		tagEnd, err := findXMLTagEnd(inner, index)
		if err != nil {
			return nil, err
		}
		token := inner[index:tagEnd]

		switch {
		case strings.HasPrefix(token, "<!--"), strings.HasPrefix(token, "<?"), strings.HasPrefix(token, "<!"):
			if depth == 0 && startIndex >= 0 {
				result = append(result, inner[startIndex:tagEnd])
				startIndex = -1
			}
		case strings.HasPrefix(token, "</"):
			depth--
			if depth < 0 {
				return nil, errors.New("invalid template xml structure")
			}
			if depth == 0 && startIndex >= 0 {
				result = append(result, inner[startIndex:tagEnd])
				startIndex = -1
			}
		case strings.HasSuffix(strings.TrimSpace(token), "/>"):
			if depth == 0 && startIndex >= 0 {
				result = append(result, inner[startIndex:tagEnd])
				startIndex = -1
			}
		default:
			depth++
		}

		index = tagEnd
	}

	if depth != 0 {
		return nil, errors.New("template xml is not balanced")
	}
	return result, nil
}

func findXMLTagEnd(text string, start int) (int, error) {
	quote := byte(0)
	for index := start; index < len(text); index++ {
		char := text[index]
		if quote != 0 {
			if char == quote {
				quote = 0
			}
			continue
		}
		if char == '"' || char == '\'' {
			quote = char
			continue
		}
		if char == '>' {
			return index + 1, nil
		}
	}
	return 0, errors.New("unterminated xml tag")
}

func xmlFragmentHasStartTag(fragment, tag string) bool {
	trimmed := strings.TrimSpace(fragment)
	prefix := "<" + tag
	if !strings.HasPrefix(trimmed, prefix) {
		return false
	}
	if len(trimmed) == len(prefix) {
		return false
	}
	switch trimmed[len(prefix)] {
	case '>', ' ', '/', '\n', '\r', '\t':
		return true
	default:
		return false
	}
}

func isWordXMLWhitespace(char byte) bool {
	switch char {
	case ' ', '\n', '\r', '\t':
		return true
	default:
		return false
	}
}

func readDocxZipEntries(data []byte) (map[string][]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte, len(reader.File))
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			return nil, err
		}
		content, readErr := io.ReadAll(handle)
		closeErr := handle.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		result[file.Name] = content
	}
	return result, nil
}

func writeDocxZipEntries(entries map[string][]byte) ([]byte, error) {
	keys := make([]string, 0, len(entries))
	for name := range entries {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, name := range keys {
		header := &zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		}
		fileWriter, err := writer.CreateHeader(header)
		if err != nil {
			_ = writer.Close()
			return nil, err
		}
		if _, err := fileWriter.Write(entries[name]); err != nil {
			_ = writer.Close()
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func parseDocxRelationships(data []byte) (docxRelationships, error) {
	var rels docxRelationships
	if err := xml.Unmarshal(data, &rels); err != nil {
		return docxRelationships{}, err
	}
	if rels.XMLNS == "" {
		rels.XMLNS = "http://schemas.openxmlformats.org/package/2006/relationships"
	}
	return rels, nil
}

func parseDocxContentTypes(data []byte) (docxContentTypes, error) {
	var types docxContentTypes
	if err := xml.Unmarshal(data, &types); err != nil {
		return docxContentTypes{}, err
	}
	if types.XMLNS == "" {
		types.XMLNS = "http://schemas.openxmlformats.org/package/2006/content-types"
	}
	return types, nil
}

func marshalXMLDocument(value any) ([]byte, error) {
	content, err := xml.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), content...), nil
}

func nextDocxRelationshipIndex(rels []docxRelationship) int {
	maxValue := 0
	for _, rel := range rels {
		if !strings.HasPrefix(rel.ID, "rId") {
			continue
		}
		value, err := strconv.Atoi(strings.TrimPrefix(rel.ID, "rId"))
		if err != nil {
			continue
		}
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func ensureDocxContentTypeDefault(types docxContentTypes, extension, contentType string) docxContentTypes {
	normalizedExtension := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(extension), "."))
	if normalizedExtension == "" || strings.TrimSpace(contentType) == "" {
		return types
	}

	for _, item := range types.Defaults {
		if strings.EqualFold(strings.TrimSpace(item.Extension), normalizedExtension) {
			return types
		}
	}

	types.Defaults = append(types.Defaults, docxContentTypeDefault{
		Extension:   normalizedExtension,
		ContentType: strings.TrimSpace(contentType),
	})
	sort.SliceStable(types.Defaults, func(i, j int) bool {
		return strings.ToLower(types.Defaults[i].Extension) < strings.ToLower(types.Defaults[j].Extension)
	})
	return types
}

func extensionByContentType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/gif":
		return "gif"
	default:
		return ""
	}
}

func normalizeContentTypeByExtension(contentType, extension string) string {
	normalizedContentType := strings.ToLower(strings.TrimSpace(contentType))
	if normalizedContentType != "" {
		return normalizedContentType
	}
	switch strings.ToLower(strings.TrimPrefix(extension, ".")) {
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	default:
		return ""
	}
}

func fileExtensionWithoutDot(name string) string {
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(strings.TrimSpace(name)), "."))
	switch extension {
	case "jpeg":
		return "jpg"
	default:
		return extension
	}
}

func mergeExportTeacherNames(snapshotTeacherName, mainTeacherName, assistants string) string {
	result := make([]string, 0, 4)
	appendUniqueTeacherNames(&result, snapshotTeacherName)
	appendUniqueTeacherNames(&result, mainTeacherName)
	appendUniqueTeacherNames(&result, assistants)
	return strings.Join(result, "、")
}

func appendUniqueTeacherNames(result *[]string, raw string) {
	for _, item := range splitExportTeacherNames(raw) {
		exists := false
		for _, current := range *result {
			if current == item {
				exists = true
				break
			}
		}
		if !exists {
			*result = append(*result, item)
		}
	}
}

func splitExportTeacherNames(raw string) []string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil
	}

	replacer := strings.NewReplacer("，", ",", "、", ",", ";", ",", "；", ",", "\n", ",")
	parts := strings.Split(replacer.Replace(text), ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func escapeXMLText(text string) string {
	var buffer bytes.Buffer
	_ = xml.EscapeText(&buffer, []byte(text))
	return buffer.String()
}
