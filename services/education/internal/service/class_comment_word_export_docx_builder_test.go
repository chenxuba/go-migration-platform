package service

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildClassCommentWordDocx(t *testing.T) {
	docxBytes, err := buildClassCommentWordDocx([]rehabRecordWordExportView{
		{
			StudentName:          "王安全",
			Gender:               "男",
			BirthDate:            "2018-01-02",
			ClassName:            "感统集体课1班",
			TeacherName:          "刘智明、李敏镐、赵子龙",
			TrainingDate:         "2026-04-13",
			TrainingTarget:       "提高孩子上课的配合能力",
			TrainingItems:        []model.RehabRecordTrainingItem{{Title: "大龙球", Content: "训练内容A"}, {Title: "羊角球", Content: "训练内容B"}},
			Performance:          "课堂表现稳定",
			Suggestion:           "持续巩固训练",
			ParentFeedback:       "家长已知悉",
			ParentSignatureImage: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4////fwAJ+wP+ONeWhwAAAABJRU5ErkJggg==",
			FeedbackDate:         "2026-04-19",
		},
	})
	if err != nil {
		t.Fatalf("buildClassCommentWordDocx failed: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("open generated docx failed: %v", err)
	}

	documentXML := readZipEntryForTest(t, reader, classCommentWordExportDocumentPath)
	relsXML := readZipEntryForTest(t, reader, classCommentWordExportDocumentRelsPath)
	contentTypesXML := readZipEntryForTest(t, reader, classCommentWordExportContentTypesPath)

	for _, expected := range []string{"王安全", "刘智明、李敏镐、赵子龙", "2026-04-13", "家长签名："} {
		if !strings.Contains(documentXML, expected) {
			t.Fatalf("generated document.xml missing %q", expected)
		}
	}
	if !strings.Contains(relsXML, classCommentWordExportImageRelationshipType) {
		t.Fatalf("generated document rels missing image relationship: %s", relsXML)
	}
	if !strings.Contains(contentTypesXML, `Extension="png"`) {
		t.Fatalf("generated content types missing png default: %s", contentTypesXML)
	}
	if readZipEntryForTest(t, reader, "word/media/rehab-signature-1.png") == "" {
		t.Fatalf("generated docx missing signature media file")
	}
}

func readZipEntryForTest(t *testing.T, reader *zip.Reader, name string) string {
	t.Helper()
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			t.Fatalf("open zip entry %s failed: %v", name, err)
		}
		defer handle.Close()

		data, err := io.ReadAll(handle)
		if err != nil {
			t.Fatalf("read zip entry %s failed: %v", name, err)
		}
		return string(data)
	}
	t.Fatalf("zip entry not found: %s", name)
	return ""
}
