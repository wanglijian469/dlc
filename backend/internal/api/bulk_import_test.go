package api

import (
	"archive/zip"
	"bytes"
	"fmt"
	"testing"
)

func TestParseXLSXReadsInlineStringsAndColumnsAfterZ(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	files := map[string]string{
		"xl/workbook.xml":            `<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="厂商信息" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml":   `<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>厂商名称*</t></is></c><c r="AA1" t="inlineStr"><is><t>服务区域</t></is></c></row><row r="2"><c r="A2" t="inlineStr"><is><t>测试厂商</t></is></c><c r="AA2" t="inlineStr"><is><t>全国</t></is></c></row></sheetData></worksheet>`,
	}
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fmt.Fprint(entry, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	sheets, err := parseXLSX(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sheet := sheets["厂商信息"]
	if len(sheet.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(sheet.Rows))
	}
	if got := sheet.Rows[1].Cells[26]; got != "全国" {
		t.Fatalf("AA2 = %q, want 全国", got)
	}
}

func TestImportSpecsJSON(t *testing.T) {
	got, err := importSpecsJSON("作业幅宽=3.6m；配套动力=80-100马力")
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"name":"作业幅宽","value":"3.6m"},{"name":"配套动力","value":"80-100马力"}]`
	if got != want {
		t.Fatalf("specs = %s, want %s", got, want)
	}
}

func TestParseOptionalBoolTreatsCollectionPlaceholdersAsEmpty(t *testing.T) {
	for _, value := range []string{"待核实", "待确认", "待补充", "暂无", "未知", "不详", "N/A", "—"} {
		parsed, err := parseOptionalBool(value)
		if err != nil {
			t.Fatalf("parseOptionalBool(%q) returned error: %v", value, err)
		}
		if parsed != nil {
			t.Fatalf("parseOptionalBool(%q) = %t, want nil", value, *parsed)
		}
	}
}

func TestCanonicalImportHeaderSupportsPreviousTemplate(t *testing.T) {
	cases := map[string]string{
		"Logo 文件名或 URL": "Logo URL",
		"厂房/门店封面图 URL":  "封面图 URL",
		"产品主图文件名或 URL":  "产品主图 URL",
		"厂商补充说明":        "厂商描述",
	}
	for input, want := range cases {
		if got := canonicalImportHeader(input); got != want {
			t.Fatalf("canonicalImportHeader(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeImportWebsiteURLAcceptsBareDomains(t *testing.T) {
	cases := map[string]string{
		"www.scaffoldwelder.com":           "https://www.scaffoldwelder.com",
		"langkepto.com":                    "https://langkepto.com",
		"vendor.example.com/products?id=1": "https://vendor.example.com/products?id=1",
		"http://vendor.example.com":        "http://vendor.example.com",
		"https://vendor.example.com":       "https://vendor.example.com",
	}
	for input, want := range cases {
		got, err := normalizeImportWebsiteURL(input)
		if err != nil {
			t.Fatalf("normalizeImportWebsiteURL(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Fatalf("normalizeImportWebsiteURL(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeImportWebsiteURLRejectsInvalidValues(t *testing.T) {
	for _, input := range []string{"not a website", "javascript://alert", "//vendor.example.com", "https://"} {
		if _, err := normalizeImportWebsiteURL(input); err == nil {
			t.Fatalf("normalizeImportWebsiteURL(%q) returned nil error", input)
		}
	}
}
