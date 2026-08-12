import { mkdirSync, writeFileSync, rmSync, copyFileSync, existsSync } from "node:fs";
import { join, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { tmpdir } from "node:os";

const root = resolve(import.meta.dirname, "..");
const output = join(root, "docs", "农机配件平台_厂商产品资料采集模板.xlsx");
const publicOutput = join(root, "frontend", "public", "templates", "农机配件平台_厂商产品资料采集模板.xlsx");
const temp = join(tmpdir(), `dalu-import-template-${Date.now()}`);

const vendorHeaders = [
  "厂商名称*", "简称", "省份", "城市", "区县", "详细地址", "官网 URL", "联系人", "联系电话", "微信/其他联系方式",
  "成立年份", "厂房面积", "员工人数", "服务优势", "主营产品", "企业简介", "年产能/供货能力", "主要设备", "资质认证",
  "是否提供加工", "加工服务", "加工材料", "加工设备", "加工能力/精度", "服务区域", "加工说明", "Logo URL", "封面图 URL", "发布状态", "推荐厂商", "平台认证", "排序",
];
const vendorExample = [
  "示例：河北某某农机配件有限公司", "某某农机", "河北省", "石家庄市", "", "河北省某县产业园", "https://vendor.example.cn", "张三", "13800000000", "微信：example",
  "2012", "12000平方米", "80人", "质量稳定、交付及时", "齿轮、链条、液压件", "企业简介示例", "年产20万套", "数控车床、热处理设备", "ISO9001",
  "是", "来图加工、来样加工", "45钢、铸铁", "数控车床、磨床", "轴类±0.02mm", "华北地区", "支持小批量定制", "https://vendor.example.cn/logo.png", "https://vendor.example.cn/cover.jpg", "隐藏", "否", "否", "100",
];
const productHeaders = [
  "产品名称*", "产品分类*", "适配机型", "产品简介", "详细说明", "关键参数（参数名=参数值）", "价格说明", "产品主图 URL", "产品图库 URL（多个用换行）", "发布状态", "热门", "推荐", "排序",
  "关联厂商名称*", "厂商产品名称", "厂商型号", "厂商适配机型", "厂商产品图片 URL", "厂商产品图库 URL（多个用换行）", "厂商描述", "厂商价格说明", "询价文案", "询价链接",
];
const productExample = [
  "示例：液压油泵总成", "液压系统配件", "拖拉机、收割机", "压力稳定，适配多种液压回路", "可根据设备型号匹配压力和接口。", "额定压力=20MPa；排量=25mL/r", "面议 / 批量报价", "https://vendor.example.cn/pump.jpg", "https://vendor.example.cn/pump-1.jpg\nhttps://vendor.example.cn/pump-2.jpg", "隐藏", "否", "否", "100",
  "河北某某农机配件有限公司", "液压油泵总成 A25", "A25", "多型号拖拉机", "https://vendor.example.cn/pump-a25.jpg", "", "支持选型和批量供货", "面议", "联系厂商", "/vendors/1",
];
const categories = ["农机易损件", "传动配件", "变速箱齿轮", "行走底盘配件", "液压系统配件", "动力发动机配件", "制动换挡配件", "电气照明配件", "收获割台配件", "播种施肥配件", "加工服务"];

const sheets = [
  {
    name: "填写说明",
    rows: [
      ["项目", "说明"],
      ["导入顺序", "先在后台“厂商信息”导入厂商，再到“配件产品”导入产品与供应关系。"],
      ["数据规则", "一行一个厂商；配件产品表一行代表一个产品与一个厂商的供应关系。同一产品有多个厂商时可重复产品名称，每行填写不同厂商。"],
      ["更新规则", "厂商按厂商名称匹配，产品按产品名称匹配，供应关系按产品+厂商匹配。已存在则更新，空白单元格保留原值。"],
      ["发布安全", "新记录未填写发布状态时默认隐藏。发布状态可填写：草稿、已发布、隐藏。"],
      ["图片规则", "图片字段填写可访问的 http/https URL 或平台已有的 /api/media/{ID} 地址；XLSX 不嵌入图片文件。"],
      ["布尔字段", "是否提供加工、推荐厂商、平台认证、热门、推荐只填写“是”或“否”。"],
      ["关键参数", "格式：参数名=参数值；参数名=参数值。多个参数可用中文分号、英文分号或换行分隔。"],
      ["示例行", "模板第 2 行是示例数据，导入器会自动跳过以“示例：”开头的行；请从第 3 行开始填写真实数据。"],
      ["校验方式", "导入会先校验整张工作表。存在错误时不会写入任何数据，并会返回工作表、行号和错误原因。"],
    ],
    widths: [20, 100],
    freeze: false,
  },
  {
    name: "厂商信息",
    rows: [vendorHeaders, vendorExample, Array(vendorHeaders.length).fill("")],
    widths: vendorHeaders.map((_, index) => index === 0 ? 32 : [5, 6, 13, 14, 15, 16, 17, 18, 19, 20, 22, 23, 24, 25, 26, 27].includes(index) ? 25 : 16),
    validations: [
      { range: "T2:T2000", values: "是,否" },
      { range: "AC2:AC2000", values: "草稿,已发布,隐藏" },
      { range: "AD2:AE2000", values: "是,否" },
    ],
  },
  {
    name: "配件产品",
    rows: [productHeaders, productExample, Array(productHeaders.length).fill("")],
    widths: productHeaders.map((_, index) => [0, 1, 13].includes(index) ? 28 : [3, 4, 5, 8, 18, 19].includes(index) ? 30 : 18),
    validations: [
      { range: "J2:J2000", values: "草稿,已发布,隐藏" },
      { range: "K2:L2000", values: "是,否" },
    ],
  },
  {
    name: "字段说明",
    rows: [
      ["字段", "可选值/填写说明"],
      ["产品分类*", categories.join("、")],
      ["发布状态", "草稿、已发布、隐藏；留空时新记录默认为隐藏"],
      ["布尔字段", "是、否"],
      ["厂商名称*", "必须与“厂商信息”中导入成功或后台已存在的厂商名称完全一致"],
      ["重复产品", "同一产品多个厂商供应时，重复填写产品名称和分类，分别填写关联厂商；产品主体只更新一次"],
      ["空白单元格", "更新已有记录时不覆盖原值"],
    ],
    widths: [24, 110],
    freeze: false,
  },
];

function escapeXml(value) {
  return String(value ?? "").replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;");
}

function columnName(index) {
  let value = index + 1;
  let result = "";
  while (value > 0) {
    value -= 1;
    result = String.fromCharCode(65 + (value % 26)) + result;
    value = Math.floor(value / 26);
  }
  return result;
}

function worksheetXml(sheet) {
  const maxColumn = columnName(Math.max(...sheet.rows.map((row) => row.length)) - 1);
  const columns = (sheet.widths || []).map((width, index) => `<col min="${index + 1}" max="${index + 1}" width="${width}" customWidth="1"/>`).join("");
  const rows = sheet.rows.map((row, rowIndex) => {
    const height = rowIndex === 0 ? 34 : rowIndex === 1 ? 58 : 24;
    const cells = row.map((value, columnIndex) => {
      const ref = `${columnName(columnIndex)}${rowIndex + 1}`;
      const style = rowIndex === 0 ? 1 : rowIndex === 1 ? 2 : 3;
      return `<c r="${ref}" t="inlineStr" s="${style}"><is><t xml:space="preserve">${escapeXml(value)}</t></is></c>`;
    }).join("");
    return `<row r="${rowIndex + 1}" ht="${height}" customHeight="1">${cells}</row>`;
  }).join("");
  const validations = (sheet.validations || []).map((item) => `<dataValidation type="list" allowBlank="1" showErrorMessage="1" errorTitle="填写值无效" error="请从下拉列表中选择" sqref="${item.range}"><formula1>&quot;${item.values}&quot;</formula1></dataValidation>`).join("");
  return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <dimension ref="A1:${maxColumn}${sheet.rows.length}"/>
  <sheetViews><sheetView workbookViewId="0">${sheet.freeze === false ? "" : '<pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/>'}</sheetView></sheetViews>
  <sheetFormatPr defaultRowHeight="20"/>
  <cols>${columns}</cols>
  <sheetData>${rows}</sheetData>
  ${sheet.freeze === false ? "" : `<autoFilter ref="A1:${maxColumn}2000"/>`}
  ${validations ? `<dataValidations count="${sheet.validations.length}">${validations}</dataValidations>` : ""}
</worksheet>`;
}

function write(relative, content) {
  const target = join(temp, relative);
  mkdirSync(resolve(target, ".."), { recursive: true });
  writeFileSync(target, content, "utf8");
}

rmSync(temp, { recursive: true, force: true });
mkdirSync(temp, { recursive: true });
mkdirSync(resolve(output, ".."), { recursive: true });
mkdirSync(resolve(publicOutput, ".."), { recursive: true });

write("[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>${sheets.map((_, index) => `<Override PartName="/xl/worksheets/sheet${index + 1}.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`).join("")}</Types>`);
write("_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`);
write("xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><bookViews><workbookView activeTab="1"/></bookViews><sheets>${sheets.map((sheet, index) => `<sheet name="${escapeXml(sheet.name)}" sheetId="${index + 1}" r:id="rId${index + 1}"/>`).join("")}</sheets></workbook>`);
write("xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">${sheets.map((_, index) => `<Relationship Id="rId${index + 1}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet${index + 1}.xml"/>`).join("")}<Relationship Id="rId${sheets.length + 1}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>`);
write("xl/styles.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fonts count="3"><font><sz val="11"/><name val="Microsoft YaHei"/></font><font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Microsoft YaHei"/></font><font><i/><sz val="10"/><color rgb="FF64748B"/><name val="Microsoft YaHei"/></font></fonts><fills count="4"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FF1559C7"/></patternFill></fill><fill><patternFill patternType="solid"><fgColor rgb="FFF1F5F9"/></patternFill></fill></fills><borders count="2"><border/><border><left style="thin"><color rgb="FFD8E1EC"/></left><right style="thin"><color rgb="FFD8E1EC"/></right><top style="thin"><color rgb="FFD8E1EC"/></top><bottom style="thin"><color rgb="FFD8E1EC"/></bottom></border></borders><cellXfs count="4"><xf fontId="0" fillId="0" borderId="0"/><xf fontId="1" fillId="2" borderId="1" applyAlignment="1"><alignment vertical="center" wrapText="1"/></xf><xf fontId="2" fillId="3" borderId="1" applyAlignment="1"><alignment vertical="top" wrapText="1"/></xf><xf fontId="0" fillId="0" borderId="1" applyAlignment="1"><alignment vertical="top" wrapText="1"/></xf></cellXfs></styleSheet>`);
sheets.forEach((sheet, index) => write(`xl/worksheets/sheet${index + 1}.xml`, worksheetXml(sheet)));

if (existsSync(output)) rmSync(output);
const packed = spawnSync("tar.exe", ["--format", "zip", "-c", "-f", output, "[Content_Types].xml", "_rels", "xl"], { cwd: temp, encoding: "utf8" });
if (packed.status !== 0) throw new Error(packed.stderr || "failed to create XLSX archive");
copyFileSync(output, publicOutput);
rmSync(temp, { recursive: true, force: true });
console.log(output);
console.log(publicOutput);
