// Command import_job_seekers imports the mock job-seeker workbook directly into
// the test database. It intentionally does not call any HTTP API.
//
// It is safe to run repeatedly: every source uid becomes one mock user and each
// source row is matched by that user plus its resume content before upserting.
package main

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	defaultWorkbook    = "scripts/mock/【整理后】求职数据.xlsx"
	mockOpenIDPrefix   = "mock-seeker:"
	nicknameCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	recruitWorkbook    = "scripts/mock/【整理后】招聘数据.xlsx"
	defaultSalaryMin   = 5000
	defaultSalaryMax   = 10000
)

type worksheet struct {
	SheetData struct {
		Rows []xlsxRow `xml:"row"`
	} `xml:"sheetData"`
}

type workbook struct {
	Sheets []workbookSheet `xml:"sheets>sheet"`
}

type workbookSheet struct {
	Name           string `xml:"name,attr"`
	RelationshipID string `xml:"id,attr"`
}

type workbookRelationships struct {
	Items []workbookRelationship `xml:"Relationship"`
}

type workbookRelationship struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type excelSheet struct {
	Name     string
	FileName string
}

type xlsxRow struct {
	Cells []xlsxCell `xml:"c"`
}

type xlsxCell struct {
	Reference string `xml:"r,attr"`
	Type      string `xml:"t,attr"`
	Value     string `xml:"v"`
	Inline    struct {
		Text string `xml:"t"`
	} `xml:"is"`
}

type sharedStrings struct {
	Items []sharedString `xml:"si"`
}

type sharedString struct {
	Text string `xml:"t"`
	Runs []struct {
		Text string `xml:"t"`
	} `xml:"r"`
}

type sourceRow struct {
	ID                  string
	SourceUID           string
	BizType             int
	Position            string
	CompanyName         string
	SocialCreditCode    string
	LegalRepresentative string
	EnterpriseAddress   string
	BusinessScope       string
	RegisteredCapital   string
	EstablishedDate     string
	ContactName         string
	Contact             string
	Description         string
	WorkContent         string
	Address             string
	AddressInfo         string
	Longitude           float64
	Latitude            float64
	Province            string
	City                string
	District            string
	ProvinceID          int
	CityID              int
	DistrictID          int
	SalaryMin           int
	SalaryMax           int
	BasicProtection     string
	SalaryBenefits      string
	AttendanceLeave     string
	RecruitNum          int
	PublishedAt         time.Time
}

func main() {
	workbookPath := flag.String("file", "", "Excel 文件路径；不传时按 -kind 选择默认文件")
	kind := flag.String("kind", "seeker", "导入类型：seeker=求职，recruit=招聘")
	sheet := flag.String("sheet", "", "要导入的 Excel 工作表名称（必填）")
	configPath := flag.String("conf", "config/test.yml", "数据库配置文件")
	execute := flag.Bool("execute", false, "实际写入数据库；默认仅预演")
	limit := flag.Int("limit", 0, "最多处理记录数，0 表示全部")
	production := flag.Bool("production", false, "允许写入生产库；必须同时提供目标用户与二次确认")
	targetUserID := flag.Int64("target-user-id", 0, "生产导入时，全部岗位关联的既有用户 ID")
	confirmProductionUserID := flag.Int64("confirm-production-user-id", 0, "生产导入二次确认，必须与 target-user-id 相同")
	flag.Parse()

	if *kind != "seeker" && *kind != "recruit" {
		fail(errors.New("-kind 仅支持 seeker 或 recruit"))
	}
	if *workbookPath == "" {
		if *kind == "recruit" {
			*workbookPath = recruitWorkbook
		} else {
			*workbookPath = defaultWorkbook
		}
	}
	rows, err := readWorkbook(*workbookPath, *kind, *sheet)
	if err != nil {
		fail(err)
	}
	if *limit > 0 && len(rows) > *limit {
		rows = rows[:*limit]
	}
	if !*execute {
		fmt.Printf("预演完成：将导入 %d 条%s信息；未连接、未写入数据库。\n", len(rows), kindLabel(*kind))
		for _, row := range rows {
			fmt.Printf("source_id=%s uid=%s %s %s%s\n", row.ID, row.SourceUID, row.Position, row.Province, row.City)
		}
		if *kind == "recruit" {
			fmt.Printf("确认后执行：bash scripts/mock/import_job_recruits.sh -conf config/test.yml -sheet %q -execute\n", *sheet)
		} else {
			fmt.Printf("确认后执行：go run ./scripts/mock/import_job_seekers.go -conf config/test.yml -sheet %q -execute\n", *sheet)
		}
		return
	}

	db, err := openDatabase(*configPath, *production)
	if err != nil {
		fail(err)
	}
	if *production && (*targetUserID <= 0 || *confirmProductionUserID != *targetUserID) {
		fail(errors.New("生产导入必须同时提供相同的 -target-user-id 和 -confirm-production-user-id"))
	}
	if *production {
		var targetUser model.User
		if err := db.First(&targetUser, *targetUserID).Error; err != nil {
			fail(fmt.Errorf("目标用户 user_id=%d 不存在: %w", *targetUserID, err))
		}
	}
	for _, row := range rows {
		var jobID, userID int64
		if err := db.Transaction(func(tx *gorm.DB) error {
			if *production {
				userID = *targetUserID
			} else {
				user, err := upsertMockUser(tx, row)
				if err != nil {
					return err
				}
				userID = user.ID
			}
			enterpriseID := int64(0)
			if row.BizType == 1 {
				enterprise, err := upsertEnterprise(tx, userID, row)
				if err != nil {
					return err
				}
				enterpriseID = enterprise.ID
			}
			job, err := upsertJob(tx, userID, enterpriseID, row)
			if err != nil {
				return err
			}
			jobID = job.ID
			return nil
		}); err != nil {
			fail(fmt.Errorf("source_id=%s 导入失败: %w", row.ID, err))
		}
		fmt.Printf("已导入：source_id=%s user_id=%d job_id=%d\n", row.ID, userID, jobID)
	}
	fmt.Printf("导入完成：共 %d 条%s信息。\n", len(rows), kindLabel(*kind))
}

func kindLabel(kind string) string {
	if kind == "recruit" {
		return "招聘"
	}
	return "求职"
}

func openDatabase(configPath string, production bool) (*gorm.DB, error) {
	expectedConfig := "test.yml"
	if production {
		expectedConfig = "prod.yml"
	}
	if filepath.Base(configPath) != expectedConfig {
		return nil, fmt.Errorf("当前模式要求 -conf 指向 %s", expectedConfig)
	}
	conf := viper.New()
	conf.SetConfigFile(configPath)
	if err := conf.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	if conf.GetString("data.db.main.driver") != "mysql" {
		return nil, errors.New("当前脚本仅支持 MySQL 测试库")
	}
	dsn := conf.GetString("data.db.main.dsn")
	if !production && !strings.Contains(strings.ToLower(dsn), "test") {
		return nil, errors.New("为防止误写生产库，测试库 DSN 必须包含 test")
	}
	if production && strings.Contains(strings.ToLower(dsn), "test") {
		return nil, errors.New("生产模式不能使用测试库 DSN")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接测试数据库失败: %w", err)
	}
	return db, nil
}

func upsertMockUser(db *gorm.DB, row sourceRow) (*model.User, error) {
	openID := mockOpenIDPrefix + row.SourceUID
	var user model.User
	err := db.Where("wechat_open_id = ?", openID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = model.User{
			WechatOpenID:          openID,
			UserCode:              mockUserCode(row.SourceUID),
			Name:                  row.ContactName,
			Phone:                 row.Contact,
			FirstAreaID:           row.ProvinceID,
			SecondAreaID:          row.CityID,
			ThirdAreaID:           row.DistrictID,
			Address:               row.Address,
			Type:                  1,
			Status:                1,
			ContactVoucherNum:     5,
			ProfileCompleteStatus: 1,
			CreateAt:              row.PublishedAt,
			UpdateAt:              row.PublishedAt,
		}
		if err := db.Create(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}
	if err != nil {
		return nil, err
	}
	updates := map[string]any{
		"name": row.ContactName, "phone": row.Contact, "first_area_id": row.ProvinceID,
		"second_area_id": row.CityID, "third_area_id": row.DistrictID, "address": row.Address,
		"update_at": time.Now(),
	}
	if err := db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func upsertEnterprise(db *gorm.DB, userID int64, row sourceRow) (*model.Enterprise, error) {
	var enterprise model.Enterprise
	err := db.Where("user_id = ? AND social_credit_code = ? AND status != ?", userID, row.SocialCreditCode, model.EnterpriseStatusDeleted).First(&enterprise).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now()
		enterprise = model.Enterprise{
			UserID: userID, Name: row.CompanyName, SocialCreditCode: row.SocialCreditCode,
			LegalRepresentative: row.LegalRepresentative, Address: row.EnterpriseAddress,
			EstablishedDate: row.EstablishedDate, BusinessPeriod: "长期有效", RegisteredCapital: row.RegisteredCapital,
			BusinessScope: row.BusinessScope, Status: model.EnterpriseStatusVerified,
			CreateAt: now, UpdateAt: now,
		}
		if err := db.Create(&enterprise).Error; err != nil {
			return nil, err
		}
		return &enterprise, nil
	}
	if err != nil {
		return nil, err
	}
	updates := map[string]any{
		"name": row.CompanyName, "legal_representative": row.LegalRepresentative,
		"address": row.EnterpriseAddress, "established_date": row.EstablishedDate,
		"business_period": "长期有效", "registered_capital": row.RegisteredCapital, "business_scope": row.BusinessScope,
		"update_at": time.Now(),
	}
	if err := db.Model(&enterprise).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &enterprise, nil
}

func upsertJob(db *gorm.DB, userID, enterpriseID int64, row sourceRow) (*model.Job, error) {
	var job model.Job
	err := db.Where("user_id = ? AND biz_type = ? AND description = ?", userID, row.BizType, row.Description).First(&job).Error
	now := time.Now()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		job = model.Job{UserID: userID}
	} else if err != nil {
		return nil, err
	}
	job.BizType = row.BizType
	job.Positions = row.Position
	job.CompanyName = row.CompanyName
	job.EnterpriseID = enterpriseID
	job.Longitude = row.Longitude
	job.Latitude = row.Latitude
	job.Address = row.Address
	job.AddressDetail = row.AddressInfo
	job.ContactPersonName = row.ContactName
	job.Contact = row.Contact
	job.Description = row.Description
	job.WorkContent = row.WorkContent
	job.Status = model.JobStatusActive
	job.FirstAreaID, job.FirstAreaDes = row.ProvinceID, row.Province
	job.SecondAreaID, job.SecondAreaDes = row.CityID, row.City
	job.ThirdAreaID, job.ThirdAreaDes = row.DistrictID, row.District
	job.SalaryMin, job.SalaryMax = row.SalaryMin, row.SalaryMax
	job.BasicProtection = row.BasicProtection
	job.SalaryBenefits = row.SalaryBenefits
	job.AttendanceLeave = row.AttendanceLeave
	job.RecruitNum = row.RecruitNum
	// 模拟导入数据按导入时刻展示，确保后台列表和信息流均排在最新位置。
	job.CreateAt = now
	job.UpdateAt = now
	job.RefreshTime = &now
	if job.ID == 0 {
		err = db.Create(&job).Error
	} else {
		err = db.Save(&job).Error
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func mockUserCode(sourceUID string) string {
	sum := sha1.Sum([]byte(mockOpenIDPrefix + sourceUID))
	return strings.ToUpper(hex.EncodeToString(sum[:])[:8])
}

// displayName keeps imported aliases consistent with the application's default
// nickname format while remaining stable when the import is run again.
func displayName(sourceName, sourceUID string) string {
	if !strings.Contains(strings.ToUpper(sourceName), "DL") {
		return sourceName
	}
	sum := sha1.Sum([]byte("nickname:" + sourceUID))
	suffix := make([]byte, 6)
	for index := range suffix {
		suffix[index] = nicknameCharacters[int(sum[index])%len(nicknameCharacters)]
	}
	return "餐饮人" + string(suffix)
}

func readWorkbook(workbookPath, kind, sheetName string) ([]sourceRow, error) {
	archive, err := zip.OpenReader(workbookPath)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	shared, err := loadSharedStrings(archive.File)
	if err != nil {
		return nil, err
	}
	sheets, err := loadWorkbookSheets(archive.File)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(sheetName) == "" {
		return nil, fmt.Errorf("请通过 -sheet 指定要导入的工作表；可选工作表：%s", sheetNames(sheets))
	}
	var selected excelSheet
	for _, sheet := range sheets {
		if sheet.Name == sheetName {
			selected = sheet
			break
		}
	}
	if selected.FileName == "" {
		return nil, fmt.Errorf("工作表 %q 不存在；可选工作表：%s", sheetName, sheetNames(sheets))
	}
	file := findZipFile(archive.File, selected.FileName)
	if file == nil {
		return nil, fmt.Errorf("工作表 %q 的数据文件不存在", sheetName)
	}
	rows, err := loadSheet(file, shared)
	if err != nil {
		return nil, fmt.Errorf("读取工作表 %q 失败: %w", sheetName, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("工作表 %q 没有数据", sheetName)
	}
	headers := cellMap(rows[0], shared)
	if _, ok := headers["岗位类别"]; !ok {
		return nil, fmt.Errorf("工作表 %q 缺少“岗位类别”列", sheetName)
	}
	return parseRows(rows[1:], headers, shared, kind)
}

func loadWorkbookSheets(files []*zip.File) ([]excelSheet, error) {
	workbookFile := findZipFile(files, "xl/workbook.xml")
	relationshipFile := findZipFile(files, "xl/_rels/workbook.xml.rels")
	if workbookFile == nil || relationshipFile == nil {
		return nil, errors.New("Excel 缺少工作表定义")
	}
	var document workbook
	if err := decodeXML(workbookFile, &document); err != nil {
		return nil, err
	}
	var relationshipDocument workbookRelationships
	if err := decodeXML(relationshipFile, &relationshipDocument); err != nil {
		return nil, err
	}
	targets := make(map[string]string, len(relationshipDocument.Items))
	for _, relationship := range relationshipDocument.Items {
		targets[relationship.ID] = pathpkg.Join("xl", relationship.Target)
	}
	result := make([]excelSheet, 0, len(document.Sheets))
	for _, sheet := range document.Sheets {
		if target := targets[sheet.RelationshipID]; target != "" {
			result = append(result, excelSheet{Name: sheet.Name, FileName: target})
		}
	}
	if len(result) == 0 {
		return nil, errors.New("Excel 不包含可读取的工作表")
	}
	return result, nil
}

func decodeXML(file *zip.File, target any) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	return xml.NewDecoder(reader).Decode(target)
}

func findZipFile(files []*zip.File, name string) *zip.File {
	for _, file := range files {
		if file.Name == name {
			return file
		}
	}
	return nil
}

func sheetNames(sheets []excelSheet) string {
	names := make([]string, 0, len(sheets))
	for _, sheet := range sheets {
		names = append(names, sheet.Name)
	}
	return strings.Join(names, "、")
}

func loadSharedStrings(files []*zip.File) ([]string, error) {
	for _, file := range files {
		if file.Name != "xl/sharedStrings.xml" {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		var document sharedStrings
		if err := xml.NewDecoder(reader).Decode(&document); err != nil {
			return nil, err
		}
		result := make([]string, len(document.Items))
		for index, item := range document.Items {
			result[index] = item.Text
			for _, run := range item.Runs {
				result[index] += run.Text
			}
		}
		return result, nil
	}
	return nil, nil
}

func loadSheet(file *zip.File, shared []string) ([]xlsxRow, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var document worksheet
	if err := xml.NewDecoder(reader).Decode(&document); err != nil {
		return nil, err
	}
	return document.SheetData.Rows, nil
}

func cellMap(row xlsxRow, shared []string) map[string]int {
	result := make(map[string]int, len(row.Cells))
	for _, cell := range row.Cells {
		result[cellValue(cell, shared)] = columnIndex(cell.Reference)
	}
	return result
}

func parseRows(rows []xlsxRow, headers map[string]int, shared []string, kind string) ([]sourceRow, error) {
	if kind == "recruit" {
		return parseRecruitRows(rows, headers, shared)
	}
	return parseSeekerRows(rows, headers, shared)
}

func parseSeekerRows(rows []xlsxRow, headers map[string]int, shared []string) ([]sourceRow, error) {
	for _, header := range []string{"id", "岗位类别", "user", "电话", "自我介绍", "省", "市", "区县", "省代码", "市代码", "区县代码"} {
		if _, ok := headers[header]; !ok {
			return nil, fmt.Errorf("Excel 缺少必需列：%s", header)
		}
	}
	salaryMinHeader := firstHeader(headers, "money_min", "薪资下限")
	if salaryMinHeader == "" {
		return nil, errors.New("Excel 缺少必需列：money_min 或 薪资下限")
	}
	salaryMaxHeader := firstHeader(headers, "money_max", "薪资上限")
	if salaryMaxHeader == "" {
		return nil, errors.New("Excel 缺少必需列：money_max 或 薪资上限")
	}
	sourceUIDHeader := firstHeader(headers, "uid")
	if sourceUIDHeader == "" {
		sourceUIDHeader = "id"
	}
	result := make([]sourceRow, 0, len(rows))
	for index, row := range rows {
		values := make(map[int]string, len(row.Cells))
		for _, cell := range row.Cells {
			values[columnIndex(cell.Reference)] = cellValue(cell, shared)
		}
		get := func(name string) string { return strings.TrimSpace(values[headers[name]]) }
		if get("id") == "" {
			continue
		}
		rowNumber := index + 2
		for _, field := range []string{sourceUIDHeader, "岗位类别", "user", "电话", "自我介绍", "省", "市", "区县", "省代码", "市代码", "区县代码", salaryMinHeader, salaryMaxHeader} {
			if get(field) == "" {
				return nil, fmt.Errorf("第 %d 行缺少 %s", rowNumber, field)
			}
		}
		publishedAt := time.Now()
		if raw := get("time"); raw != "" {
			parsed, err := time.ParseInLocation("2006-01-02 15:04:05", raw, time.Local)
			if err != nil {
				return nil, fmt.Errorf("第 %d 行 time 格式不正确: %w", rowNumber, err)
			}
			publishedAt = parsed
		}
		provinceID, err := strconv.Atoi(get("省代码"))
		if err != nil {
			return nil, fmt.Errorf("第 %d 行省代码必须是整数: %w", rowNumber, err)
		}
		cityID, err := strconv.Atoi(get("市代码"))
		if err != nil {
			return nil, fmt.Errorf("第 %d 行市代码必须是整数: %w", rowNumber, err)
		}
		districtID, err := strconv.Atoi(get("区县代码"))
		if err != nil {
			return nil, fmt.Errorf("第 %d 行区县代码必须是整数: %w", rowNumber, err)
		}
		salaryMin, err := strconv.Atoi(get(salaryMinHeader))
		if err != nil {
			return nil, fmt.Errorf("第 %d 行 %s 必须是整数: %w", rowNumber, salaryMinHeader, err)
		}
		salaryMax, err := strconv.Atoi(get(salaryMaxHeader))
		if err != nil {
			return nil, fmt.Errorf("第 %d 行 %s 必须是整数: %w", rowNumber, salaryMaxHeader, err)
		}
		salaryMin, salaryMax = normalizeSalaryRange(salaryMin, salaryMax)
		sourceUID := get(sourceUIDHeader)
		entry := sourceRow{
			ID: get("id"), SourceUID: sourceUID, BizType: 2, Position: get("岗位类别"), ContactName: displayName(get("user"), sourceUID), Contact: get("电话"),
			Description: get("自我介绍"), Address: get("address"), AddressInfo: get("address_detail"), Province: get("省"), City: get("市"), District: get("区县"),
			WorkContent: get("自我介绍"), ProvinceID: provinceID, CityID: cityID, DistrictID: districtID, SalaryMin: salaryMin, SalaryMax: salaryMax, PublishedAt: publishedAt,
		}
		result = append(result, entry)
	}
	return result, nil
}

func firstHeader(headers map[string]int, candidates ...string) string {
	for _, candidate := range candidates {
		if _, ok := headers[candidate]; ok {
			return candidate
		}
	}
	return ""
}

func parseRecruitRows(rows []xlsxRow, headers map[string]int, shared []string) ([]sourceRow, error) {
	// 2026.8.6 and 2026.8.7 use different export field names. Normalize them
	// here so both sheets produce the same sourceRow.
	idHeader := firstHeader(headers, "id", "uid")
	sourceUIDHeader := firstHeader(headers, "uid")
	positionHeader := firstHeader(headers, "岗位类别")
	companyNameHeader := firstHeader(headers, "企业名称")
	contactNameHeader := firstHeader(headers, "user")
	contactHeader := firstHeader(headers, "电话")
	descriptionHeader := firstHeader(headers, "岗位要求", "工作要求")
	workContentHeader := firstHeader(headers, "工作内容")
	// Both source sheets label latitude as “经度” and longitude as “纬度/维度”.
	// Store them according to the actual numeric values, not the source labels.
	latitudeHeader := firstHeader(headers, "经度")
	longitudeHeader := firstHeader(headers, "纬度", "维度")
	provinceHeader := firstHeader(headers, "省")
	cityHeader := firstHeader(headers, "市")
	districtHeader := firstHeader(headers, "区县")
	provinceIDHeader := firstHeader(headers, "省代码")
	cityIDHeader := firstHeader(headers, "市代码")
	districtIDHeader := firstHeader(headers, "区县代码")
	salaryMinHeader := firstHeader(headers, "money_min", "薪资下限")
	salaryMaxHeader := firstHeader(headers, "money_max", "薪资上限")
	recruitNumHeader := firstHeader(headers, "招聘人数")
	socialCreditCodeHeader := firstHeader(headers, "统一社会信用代码")
	establishedDateHeader := firstHeader(headers, "注册日期")
	for _, required := range []struct {
		name   string
		header string
	}{
		{"来源记录标识（id 或 uid）", idHeader}, {"uid", sourceUIDHeader}, {"岗位类别", positionHeader},
		{"企业名称", companyNameHeader}, {"user", contactNameHeader}, {"电话", contactHeader},
		{"岗位要求或工作要求", descriptionHeader}, {"工作内容", workContentHeader},
		{"经度", latitudeHeader}, {"纬度或维度", longitudeHeader}, {"省", provinceHeader},
		{"市", cityHeader}, {"区县", districtHeader}, {"省代码", provinceIDHeader},
		{"市代码", cityIDHeader}, {"区县代码", districtIDHeader},
		{"money_min 或薪资下限", salaryMinHeader}, {"money_max 或薪资上限", salaryMaxHeader},
		{"招聘人数", recruitNumHeader}, {"统一社会信用代码", socialCreditCodeHeader}, {"注册日期", establishedDateHeader},
	} {
		if required.header == "" {
			return nil, fmt.Errorf("Excel 缺少招聘导入所需列：%s", required.name)
		}
	}
	legalRepresentativeHeader := firstHeader(headers, "企业法人", "法人代表")
	registeredCapitalHeader := firstHeader(headers, "注册资本", "注册资产")
	enterpriseAddressHeader := firstHeader(headers, "企业地址")
	businessScopeHeader := firstHeader(headers, "经营范围")
	addressHeader := firstHeader(headers, "address")
	addressDetailHeader := firstHeader(headers, "address_detail")
	basicProtectionHeader := firstHeader(headers, "基础保障", "basic_protection")
	salaryBenefitsHeader := firstHeader(headers, "薪酬福利", "salary_benefits")
	attendanceLeaveHeader := firstHeader(headers, "考勤休假", "attendance_leave")
	result := make([]sourceRow, 0, len(rows))
	for index, row := range rows {
		values := make(map[int]string, len(row.Cells))
		for _, cell := range row.Cells {
			values[columnIndex(cell.Reference)] = cellValue(cell, shared)
		}
		get := func(header string) string {
			if header == "" {
				return ""
			}
			return strings.TrimSpace(values[headers[header]])
		}
		if get(idHeader) == "" {
			continue
		}
		rowNumber := index + 2
		for _, field := range []string{sourceUIDHeader, positionHeader, companyNameHeader, contactNameHeader, contactHeader, descriptionHeader, workContentHeader, latitudeHeader, longitudeHeader, provinceHeader, cityHeader, districtHeader, provinceIDHeader, cityIDHeader, districtIDHeader, salaryMinHeader, salaryMaxHeader, recruitNumHeader, socialCreditCodeHeader, establishedDateHeader} {
			if get(field) == "" {
				return nil, fmt.Errorf("第 %d 行缺少 %s", rowNumber, field)
			}
		}
		provinceID, err := numberToInt(get(provinceIDHeader), provinceIDHeader, rowNumber)
		if err != nil {
			return nil, err
		}
		cityID, err := numberToInt(get(cityIDHeader), cityIDHeader, rowNumber)
		if err != nil {
			return nil, err
		}
		districtID, err := numberToInt(get(districtIDHeader), districtIDHeader, rowNumber)
		if err != nil {
			return nil, err
		}
		salaryMin, err := numberToInt(get(salaryMinHeader), salaryMinHeader, rowNumber)
		if err != nil {
			return nil, err
		}
		salaryMax, err := numberToInt(get(salaryMaxHeader), salaryMaxHeader, rowNumber)
		if err != nil {
			return nil, err
		}
		salaryMin, salaryMax = normalizeSalaryRange(salaryMin, salaryMax)
		recruitNum, err := numberToInt(get(recruitNumHeader), recruitNumHeader, rowNumber)
		if err != nil {
			return nil, err
		}
		longitude, err := strconv.ParseFloat(get(longitudeHeader), 64)
		if err != nil {
			return nil, fmt.Errorf("第 %d 行 %s 必须是数字: %w", rowNumber, longitudeHeader, err)
		}
		latitude, err := strconv.ParseFloat(get(latitudeHeader), 64)
		if err != nil {
			return nil, fmt.Errorf("第 %d 行 %s 必须是数字: %w", rowNumber, latitudeHeader, err)
		}
		establishedDate, err := registrationDate(get(establishedDateHeader))
		if err != nil {
			return nil, fmt.Errorf("第 %d 行 %s 格式不正确: %w", rowNumber, establishedDateHeader, err)
		}
		sourceUID := get(sourceUIDHeader)
		result = append(result, sourceRow{
			ID: get(idHeader), SourceUID: sourceUID, BizType: 1, Position: get(positionHeader), CompanyName: get(companyNameHeader), SocialCreditCode: get(socialCreditCodeHeader),
			LegalRepresentative: get(legalRepresentativeHeader), EnterpriseAddress: get(enterpriseAddressHeader), BusinessScope: get(businessScopeHeader), RegisteredCapital: get(registeredCapitalHeader), EstablishedDate: establishedDate,
			ContactName: displayName(get(contactNameHeader), sourceUID), Contact: get(contactHeader), Description: get(descriptionHeader), WorkContent: get(workContentHeader),
			Address: get(addressHeader), AddressInfo: get(addressDetailHeader), Longitude: longitude, Latitude: latitude,
			Province: get(provinceHeader), City: get(cityHeader), District: get(districtHeader), ProvinceID: provinceID, CityID: cityID, DistrictID: districtID,
			SalaryMin: salaryMin, SalaryMax: salaryMax,
			BasicProtection: normalizeBenefits(get(basicProtectionHeader)),
			SalaryBenefits:  normalizeBenefits(get(salaryBenefitsHeader)),
			AttendanceLeave: normalizeBenefits(get(attendanceLeaveHeader)),
			RecruitNum:      recruitNum, PublishedAt: time.Now(),
		})
	}
	return result, nil
}

// normalizeBenefits converts Excel's human-readable benefit separator into the
// comma-separated format used by the job table and existing API handlers.
func normalizeBenefits(value string) string {
	items := strings.FieldsFunc(value, func(character rune) bool {
		switch character {
		case '、', '，', ',', '；', ';', '\n', '\r':
			return true
		default:
			return false
		}
	})
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return strings.Join(result, ",")
}

// normalizeSalaryRange supplies a presentable salary range for source rows that
// use zero as an unspecified salary value.
func normalizeSalaryRange(minimum, maximum int) (int, int) {
	if minimum == 0 || maximum == 0 {
		return defaultSalaryMin, defaultSalaryMax
	}
	return minimum, maximum
}

func registrationDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "无" {
		return time.Now().Format("2006-01-02"), nil
	}
	for _, layout := range []string{"2006-01-02", "2006/1/2", "2006.1.2", "2006年01月02日", "2006年1月2日"} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed.Format("2006-01-02"), nil
		}
	}
	serial, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", err
	}
	return time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(serial)).Format("2006-01-02"), nil
}

func numberToInt(value, field string, rowNumber int) (int, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed != float64(int(parsed)) {
		return 0, fmt.Errorf("第 %d 行 %s 必须是整数", rowNumber, field)
	}
	return int(parsed), nil
}

func cellValue(cell xlsxCell, shared []string) string {
	if cell.Type == "inlineStr" {
		return cell.Inline.Text
	}
	if cell.Type == "s" {
		index, err := strconv.Atoi(cell.Value)
		if err == nil && index >= 0 && index < len(shared) {
			return shared[index]
		}
	}
	return cell.Value
}

func columnIndex(reference string) int {
	index := 0
	for _, character := range reference {
		if character < 'A' || character > 'Z' {
			break
		}
		index = index*26 + int(character-'A'+1)
	}
	return index - 1
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "导入失败：", err)
	os.Exit(1)
}
