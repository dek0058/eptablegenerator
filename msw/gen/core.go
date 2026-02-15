package gen

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"eptablegenerator/msw/config"
	"eptablegenerator/table"
	"eptablegenerator/table/xlsx"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type variableType struct {
	Name    string
	Type    string
	Default string
}

type tableDocument struct {
	Name          string
	TableContent  string
	RecordContent string
	Rows          [][]string
}

type userDataset struct {
	Id             string       `json:"Id"`
	GameId         string       `json:"GameId"`
	EntryKey       string       `json:"EntryKey"`
	ContentType    string       `json:"ContentType"`
	Content        string       `json:"Content"`
	Usage          int          `json:"Usage"`
	UsePublish     int          `json:"UsePublish"`
	UseService     int          `json:"UseService"`
	CoreVersion    string       `json:"CoreVersion"`
	StudioVersion  string       `json:"StudioVersion"`
	DynamicLoading int          `json:"DynamicLoading"`
	ContentProto   contentProto `json:"ContentProto"`
}

type contentProto struct {
	Use  string   `json:"Use"`
	Json jsonType `json:"Json"`
}

type jsonType struct {
	Name              string `json:"name"`
	Id                string `json:"id"`
	Serveronly        bool   `json:"serveronly"`
	SyncDataSetWebUrl string `json:"syncDataSetWebUrl"`
	Dynamicloading    int    `json:"dynamicloading"`
}

func isFileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil && !info.IsDir()
}

func NewUserDataset(name, uuid, coreVersions string) userDataset {
	return userDataset{
		Id:             "",
		GameId:         "",
		EntryKey:       "userdataset://" + uuid,
		ContentType:    "x-mod/userdataset",
		Content:        "",
		Usage:          0,
		UsePublish:     1,
		UseService:     0,
		CoreVersion:    coreVersions,
		StudioVersion:  "0.1.0.0",
		DynamicLoading: 0,
		ContentProto: contentProto{
			Use: "Json",
			Json: jsonType{
				Name:              name,
				Id:                uuid,
				Serveronly:        false,
				SyncDataSetWebUrl: "",
				Dynamicloading:    0,
			},
		},
	}
}

func generate(file string, sheetName string, data [][]string) (tableDocument, error) {
	if len(data) < 3 {
		log.Printf("Sheet '%s' in file '%s' has insufficient data, skipping.\n", sheetName, file)
		return tableDocument{}, errors.New("insufficient data")
	}

	scriptName := fmt.Sprintf("%sRecord", sheetName)

	headers := data[0]
	types := data[1]
	attributes := data[2]

	colmnCount := len(headers)

	if colmnCount != len(types) || colmnCount != len(attributes) {
		log.Printf("Sheet '%s' in file '%s' has mismatched header, type, and attribute counts, skipping.\n", sheetName, file)
		return tableDocument{}, errors.New("mismatched header, type, and attribute counts")
	}

	doc := tableDocument{
		Name: sheetName,
		Rows: [][]string{},
	}

	variableTypes := []*variableType{}
	indexKeyType := ""
	indexKeyName := ""

	// msw 타입을 사용
	// https://maplestoryworlds-creators.nexon.com/ko/docs?postId=208
	for i, v := range types {
		if strings.ToLower(attributes[i]) == "design" {
			continue
		} else if strings.ToLower(attributes[i]) == "key" {
			indexKeyType = v
			indexKeyName = headers[i]
		}

		header := headers[i]
		var newVar *variableType

		switch v {
		case "boolean":
			newVar = &variableType{Name: header, Type: "boolean", Default: "false"}

		case "integer":
			newVar = &variableType{Name: header, Type: "integer", Default: "0"}

		case "number":
			newVar = &variableType{Name: header, Type: "number", Default: "0.0"}

		case "string":
			newVar = &variableType{Name: header, Type: "string", Default: `""`}

		default:
			continue
		}

		variableTypes = append(variableTypes, newVar)
	}

	if indexKeyType == "" || indexKeyName == "" {
		log.Printf("Sheet '%s' in file '%s' does not have a valid key column, skipping.\n", sheetName, file)
		return tableDocument{}, errors.New("missing valid key column")
	}

	{
		var content strings.Builder
		content.WriteString("---@description \"자동 생성된 테이블 입니다. 수정하지 마세요 (Derivative: " + sheetName + ")\"\n")
		content.WriteString("@Struct\n")
		fmt.Fprintf(&content, "script %s\n", scriptName)
		content.WriteString("\n")

		for _, vType := range variableTypes {
			if vType.Type == "unknown" {
				continue
			}
			fmt.Fprintf(&content, "\tproperty %s %s = %s\n", vType.Type, vType.Name, vType.Default)
		}

		content.WriteString("\n")
		content.WriteString("end\n")

		doc.RecordContent = content.String()
	}

	{
		var content strings.Builder
		content.WriteString("---@description \"자동 생성된 테이블 입니다. 수정하지 마세요 (Derivative: " + sheetName + ")\"\n")
		content.WriteString("@Struct\n")
		fmt.Fprintf(&content, "script %sTable\n", sheetName)
		content.WriteString("\n")
		content.WriteString("\tproperty table records = {}\n")
		content.WriteString("\n")

		content.WriteString("\t---@description \"테이블을 로드 합니다\"\n")
		content.WriteString("\tmethod void Load()\n")

		fmt.Fprintf(&content, "\t\tlocal userDataset = _DataService:GetTable(\"%s\")\n", sheetName+"Table")
		content.WriteString("\t\tlocal rowCount = userDataset:GetRowCount()\n")
		content.WriteString("\n")

		content.WriteString("\t\tfor row = 1, rowCount do\n")

		fmt.Fprintf(&content, "\t\t\tlocal record = %sRecord()\n", sheetName)
		content.WriteString("\t\t\tlocal cell = \"\"\n")

		for _, vType := range variableTypes {
			switch vType.Type {
			case "boolean":
				fmt.Fprintf(&content, "\t\t\tcell = userDataset:GetCell(row, \"%s\")\n", vType.Name)
				fmt.Fprintf(&content, "\t\t\trecord.%s = (string.lower(cell) == \"true\")\n", vType.Name)

			case "integer":
				fmt.Fprintf(&content, "\t\t\tcell = userDataset:GetCell(row, \"%s\")\n", vType.Name)
				fmt.Fprintf(&content, "\t\t\trecord.%s = tonumber(cell) or 0\n", vType.Name)

			case "number":
				fmt.Fprintf(&content, "\t\t\tcell = userDataset:GetCell(row, \"%s\")\n", vType.Name)
				fmt.Fprintf(&content, "\t\t\trecord.%s = tonumber(cell) or 0.0\n", vType.Name)

			case "string":
				fmt.Fprintf(&content, "\t\t\trecord.%s = userDataset:GetCell(row, \"%s\")\n", vType.Name, vType.Name)

			default:
				continue
			}
		}

		fmt.Fprintf(&content, "\t\t\tself.records[record.%s] = record\n", indexKeyName)
		content.WriteString("\t\tend\n")
		content.WriteString("\tend\n")
		content.WriteString("\n")

		content.WriteString("\t---@description \"테이블에서 레코드를 가져옵니다\"\n")
		fmt.Fprintf(&content, "\tmethod %sRecord GetRecord(%s key)\n", sheetName, indexKeyType)
		content.WriteString("\t\treturn self.records[key]\n")
		content.WriteString("\tend\n")

		content.WriteString("\n")
		content.WriteString("end\n")

		doc.TableContent = content.String()
	}

	// csv 데이터 생성
	for i, row := range data {
		if i < 3 {
			continue
		}

		newDatas := []string{}
		for j := range colmnCount {
			if strings.ToLower(attributes[j]) == "design" {
				continue
			}

			newDatas = append(newDatas, row[j])
		}

		doc.Rows = append(doc.Rows, newDatas)
	}

	return doc, nil
}

func createMlua(createdPath string, doc tableDocument) error {
	{
		path := filepath.Join(createdPath, doc.Name+"Record.mlua")
		h, err := os.Create(path)
		if err != nil {
			log.Printf("Failed to create file '%s': %v\n", path, err)
			return err
		}
		defer h.Close()

		if _, err := h.WriteString(doc.RecordContent); err != nil {
			log.Printf("Failed to write to file '%s': %v\n", path, err)
			return err
		}

		log.Printf("Generated mLua file: %s\n", path)
	}

	{
		path := filepath.Join(createdPath, doc.Name+"Table.mlua")
		h, err := os.Create(path)
		if err != nil {
			log.Printf("Failed to create file '%s': %v\n", path, err)
			return err
		}
		defer h.Close()

		if _, err := h.WriteString(doc.TableContent); err != nil {
			log.Printf("Failed to write to file '%s': %v\n", path, err)
			return err
		}

		log.Printf("Generated mLua file: %s\n", path)
	}

	return nil
}

func createCSV(createdPath string, version string, doc tableDocument) error {
	uuid := make([]byte, 16)

	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		log.Printf("Failed to generate UUID for document '%s': %v\n", doc.Name, err)
		return err
	}

	// 파일이 존재하지 않을 때만 userdataset 생성
	if !isFileExists(filepath.Join(createdPath, doc.Name+"Table.userdataset")) {
		uuid[6] = (uuid[6] & 0x0f) | 0x40
		uuid[8] = (uuid[8] & 0x3f) | 0x80
		uuidStr := fmt.Sprintf("%x-%x-%x-%x-%x",
			uuid[0:4],   // 8자리
			uuid[4:6],   // 4자리
			uuid[6:8],   // 4자리 (버전 포함)
			uuid[8:10],  // 4자리 (variant 포함)
			uuid[10:16], // 12자리
		)

		userDataset := NewUserDataset(doc.Name+"Table", uuidStr, version)
		jsonData, err := json.MarshalIndent(userDataset, "", "    ")
		if err != nil {
			log.Printf("Failed to marshal user dataset for document '%s': %v\n", doc.Name, err)
			return err
		}

		userDatasetFile := filepath.Join(createdPath, doc.Name+"Table.userdataset")
		if err := os.WriteFile(userDatasetFile, jsonData, 0644); err != nil {
			log.Printf("Failed to write user dataset file '%s': %v\n", userDatasetFile, err)
			return err
		}
	}

	path := filepath.Join(createdPath, doc.Name+"Table.csv")
	h, err := os.Create(path)
	if err != nil {
		log.Printf("Failed to create file '%s': %v\n", path, err)
		return err
	}
	defer h.Close()

	writer := csv.NewWriter(h)
	defer writer.Flush()

	if err := writer.WriteAll(doc.Rows); err != nil {
		log.Printf("Failed to write to file '%s': %v\n", path, err)
		return err
	}

	log.Printf("Generated CSV file: %s\n", path)

	return nil
}

func Generate(c *config.Config) error {
	if c == nil {
		return errors.New("config is nil")
	}

	mswData, err := os.ReadFile(c.MswConfig)
	if err != nil {
		return fmt.Errorf("failed to read MSW config file: %w", err)
	}

	var mswConfig config.MswConfig
	if err := json.Unmarshal(mswData, &mswConfig); err != nil {
		return fmt.Errorf("failed to parse MSW config file: %w", err)
	}

	files, err := table.FindXLSX(c.SourceDir)
	if err != nil {
		return err
	}

	tableDocuments := []tableDocument{}

	for _, file := range files {
		x := xlsx.NewXLSX(file)
		if x == nil {
			log.Println("Failed to load XLSX file:", file)
			continue
		}

		for sheetName, data := range x.Data {
			doc, err := generate(file, sheetName, data)
			if err != nil {
				log.Printf("Error generating document for sheet '%s' in file '%s': %v\n", sheetName, file, err)
				continue
			}

			tableDocuments = append(tableDocuments, doc)
		}
	}

	for _, doc := range tableDocuments {
		if err := createMlua(c.DestDir, doc); err != nil {
			log.Printf("Error creating mLua file for document '%s': %v\n", doc.Name, err)
			continue
		}
		if err := createCSV(c.CsvDir, mswConfig.CoreVersion, doc); err != nil {
			log.Printf("Error creating CSV file for document '%s': %v\n", doc.Name, err)
			continue
		}
	}

	return nil
}
