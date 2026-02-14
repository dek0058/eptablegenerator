package gen

import (
	"eptablegenerator/table"
	"eptablegenerator/table/xlsx"
	"eptablegenerator/ueproject/config"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

type sheetType interface {
	GetSheetName() string
	Generate() (string, []string, []string, error)
	Optional()
}

type defaultSheetType struct {
	ProjectName   string
	SheetName     string
	Data          *[][]string
	OptionalFiles []string
}

func (d *defaultSheetType) GetSheetName() string {
	return d.SheetName
}

func (d *defaultSheetType) Optional() {}

type structType struct {
	defaultSheetType
}

func (s *structType) Generate() (string, []string, []string, error) {
	var content string
	forwardContent := []string{}
	include := []string{}

	if len(*s.Data) < 2 {
		return content, forwardContent, include, errors.New("data is not enough")
	}

	// NOTE. 첫번 째 헤더 이름, 두번 째 값 타입
	header := (*s.Data)[0]
	types := (*s.Data)[1]

	type VariableType struct {
		Type    string
		Default string
	}

	variables := []VariableType{}
	for _, v := range types {
		switch {
		case v == "bool":
			variables = append(variables, VariableType{"bool", "false"})

		case v == "int32":
			variables = append(variables, VariableType{"int32", "INDEX_NONE"})

		case v == "int64":
			variables = append(variables, VariableType{"int64", "INDEX_NONE"})

		case v == "float32":
			variables = append(variables, VariableType{"float", "INDEX_NONE"})

		case v == "float64":
			variables = append(variables, VariableType{"double", "INDEX_NONE"})

		case v == "FString":
			variables = append(variables, VariableType{"FString", ""})

		case v == "FText":
			variables = append(variables, VariableType{"FText", ""})

		case strings.HasPrefix(v, "TArray<") && strings.HasSuffix(v, ">"):
			variables = append(variables, VariableType{"TArray<" + v[7:len(v)-1] + ">", ""})

		case strings.HasPrefix(v, "TMap<") && strings.HasSuffix(v, ">"):
			variables = append(variables, VariableType{"TMap<" + v[5:len(v)-1] + ">", ""})

		case strings.HasPrefix(v, "TSet<") && strings.HasSuffix(v, ">"):
			variables = append(variables, VariableType{"TSet<" + v[5:len(v)-1] + ">", ""})

		case strings.HasPrefix(v, "Enum<") && strings.HasSuffix(v, ">"):
			variables = append(variables, VariableType{v[5 : len(v)-1], "static_cast<" + v[5:len(v)-1] + ">(0)"})
			forwardContent = append(forwardContent, "enum class "+v[5:len(v)-1]+" : uint8")

		case strings.HasPrefix(v, "Class<") && strings.HasSuffix(v, ">"):
			variables = append(variables, VariableType{"TSoftClassPtr<" + v[6:len(v)-1] + ">", ""})
			forwardContent = append(forwardContent, "class "+v[6:len(v)-1])

		case strings.HasPrefix(v, "Asset<") && strings.HasSuffix(v, ">"):
			variables = append(variables, VariableType{"TSoftObjectPtr<" + v[6:len(v)-1] + ">", ""})
			forwardContent = append(forwardContent, "class "+v[6:len(v)-1])

		default:
			variables = append(variables, VariableType{})
		}
	}

	include = append(include, "Engine/DataTable.h")

	projectName := strings.ToUpper(s.ProjectName)
	if projectName != "" {
		projectName = fmt.Sprintf("%s_API ", projectName)
	}

	content += "USTRUCT(BlueprintType)\n"
	content += fmt.Sprintf("struct %sF%sTableRow : public FTableRowBase\n", projectName, s.SheetName)
	content += "{\n"
	content += "\tGENERATED_BODY()\n"
	content += "\n"

	duplicate := map[string]any{}

	varCount := len(variables)
	for i, name := range header {
		if i >= varCount {
			break
		}

		v := variables[i]
		if v.Type == "" || name == "" {
			continue
		}

		if _, ok := duplicate[name]; ok {
			continue
		}
		duplicate[name] = nil

		content += "\tUPROPERTY(EditAnywhere, BlueprintReadWrite)\n"
		if v.Default == "" {
			content += fmt.Sprintf("\t%s %s;\n", v.Type, name)
		} else {
			content += fmt.Sprintf("\t%s %s = %s;\n", v.Type, name, v.Default)
		}
		content += "\n"
	}

	content += "};\n"
	content += "\n"

	return content, forwardContent, include, nil
}

type enumType struct {
	defaultSheetType
}

func (e *enumType) Generate() (string, []string, []string, error) {
	var content string
	forwardContent := []string{}
	include := []string{}

	if len(*e.Data) < 2 {
		return content, forwardContent, include, errors.New("data is not enough")
	}

	values := []string{}
	for _, data := range (*e.Data)[2:] {
		if len(data) < 2 {
			continue
		}

		value, err := strconv.Atoi(data[0])
		name := data[1]

		var displayName string
		if len(data) >= 3 {
			displayName = data[2]
		}

		if err != nil {
			println("value is not int: " + e.SheetName)
			continue
		}

		if name == "" {
			println("name is empty: " + e.SheetName)
			continue
		}

		r := fmt.Sprintf("\t%s = %d", name, value)
		if displayName != "" {
			r += fmt.Sprintf(" UMETA(DisplayName = \"%s\")", displayName)
		}

		values = append(values, r)
	}

	include = append(include, "Misc/EnumRange.h")

	content += "UENUM(BlueprintType)\n"
	content += fmt.Sprintf("enum class %s : uint8", e.SheetName)
	content += "{\n"

	for _, v := range values {
		content += v + ",\n"
	}

	content += "\tMax UMETA(Hidden)\n"
	content += "};\n"
	content += fmt.Sprintf("ENUM_RANGE_BY_COUNT(%s, %s::Max)\n", e.SheetName, e.SheetName)
	content += "\n"

	return content, forwardContent, include, nil
}

type ConstType struct {
	defaultSheetType
}

func (c *ConstType) Generate() (string, []string, []string, error) {
	var content string
	forwardContent := []string{}
	include := []string{}

	include = append(include, "Engine/DeveloperSettings.h")

	configName := c.ProjectName
	if configName == "" {
		configName = "Game"
	}

	projectName := strings.ToUpper(c.ProjectName)
	if projectName != "" {
		projectName = fmt.Sprintf("%s_API ", projectName)
	}

	content += fmt.Sprintf("UCLASS(config = %s, defaultconfig)\n", configName)
	content += fmt.Sprintf("class %sU%sSettings : public UDeveloperSettings\n", projectName, c.SheetName)
	content += "{\n"
	content += "\tGENERATED_BODY()\n"
	content += "\n"
	content += "public:\n"

	for _, data := range (*c.Data)[1:] {
		if len(data) < 3 {
			continue
		}

		n := data[0]
		t := data[1]

		if n == "" || t == "" {
			println("name or type is empty: " + c.SheetName)
			continue
		}

		r := "\tUPROPERTY(Config, EditAnywhere, BlueprintReadOnly, Category = \"Table\")\n"

		switch {
		case t == "bool":
			r += fmt.Sprintf("\tbool %s;\n", n)

		case t == "int32":
			r += fmt.Sprintf("\tint32 %s;\n", n)

		case t == "int64":
			r += fmt.Sprintf("\tint64 %s;\n", n)

		case t == "float32":
			r += fmt.Sprintf("\tfloat %s;\n", n)

		case t == "float64":
			r += fmt.Sprintf("\tdouble %s;\n", n)

		case t == "FString":
			r += fmt.Sprintf("\tFString %s;\n", n)

		case t == "FText":
			r += fmt.Sprintf("\tFText %s;\n", n)

		case strings.HasPrefix(t, "TArray<") && strings.HasSuffix(t, ">"):
			r += fmt.Sprintf("\tTArray<%s> %s;\n", t[7:len(t)-1], n)

		case strings.HasPrefix(t, "TMap<") && strings.HasSuffix(t, ">"):
			r += fmt.Sprintf("\tTMap<%s> %s;\n", t[5:len(t)-1], n)

		case strings.HasPrefix(t, "TSet<") && strings.HasSuffix(t, ">"):
			r += fmt.Sprintf("\tTSet<%s> %s;\n", t[5:len(t)-1], n)

		case strings.HasPrefix(t, "Enum<") && strings.HasSuffix(t, ">"):
			r += fmt.Sprintf("\t%s %s;\n", t[5:len(t)-1], n)
			forwardContent = append(forwardContent, "enum class "+t[5:len(t)-1]+" : uint8")

		case strings.HasPrefix(t, "Class<") && strings.HasSuffix(t, ">"):
			r += fmt.Sprintf("\tTSoftClassPtr<%s> %s;\n", t[6:len(t)-1], n)
			forwardContent = append(forwardContent, "class "+t[6:len(t)-1])

		case strings.HasPrefix(t, "Asset<") && strings.HasSuffix(t, ">"):
			r += fmt.Sprintf("\tTSoftObjectPtr<%s> %s;\n", t[6:len(t)-1], n)
			forwardContent = append(forwardContent, "class "+t[6:len(t)-1])

		default:
			continue
		}

		content += r
		content += "\n"
	}

	content += "};\n"
	content += "\n"

	return content, forwardContent, include, nil
}

func (c *ConstType) Optional() {
	conifgFileName := "Default" + c.ProjectName + ".ini"
	var cfgPath string
	var cfg *ini.File
	var err error

	for _, v := range c.OptionalFiles {
		if strings.Contains(v, conifgFileName) {
			cfg, err = ini.Load(v)
			cfgPath = v
			break
		}
	}

	if cfg == nil || err != nil {
		return
	}

	section := "/Script/" + c.ProjectName + "." + c.SheetName + "Settings"

	for _, data := range (*c.Data)[1:] {
		if len(data) < 3 {
			continue
		}

		n := data[0]
		v := data[2]

		cfg.Section(section).Key(n).SetValue(v)
	}

	cfg.SaveTo(cfgPath)
}

func Generate(c *config.Config) error {
	if c == nil {
		return errors.New("config is nil")
	}

	files, err := table.FindXLSX(c.SourceDir)
	if err != nil {
		return err
	}

	m := map[string][]sheetType{}
	for _, file := range files {
		x := xlsx.NewXLSX(file)
		fileName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		for sheetName, data := range x.Data {
			switch {

			// 구조체 타입
			case strings.HasPrefix(sheetName, "!"):
				// NOTE. 첫번 째 헤더 이름, 두번 째 값 타입
				sliceData := data[:2]
				st := &structType{defaultSheetType{
					ProjectName:   c.ProjectName,
					SheetName:     sheetName[1:],
					Data:          &sliceData,
					OptionalFiles: c.OptionalFiles,
				}}
				m[fileName] = append(m[fileName], st)

				// 열거형 타입
			case strings.HasPrefix(sheetName, "@"):
				et := &enumType{defaultSheetType{
					ProjectName:   c.ProjectName,
					SheetName:     sheetName[1:],
					Data:          &data,
					OptionalFiles: c.OptionalFiles,
				}}
				m[fileName] = append(m[fileName], et)

				// 글로벌 매직 변수 타입
			case strings.HasPrefix(sheetName, "#"):
				ct := &ConstType{defaultSheetType{
					ProjectName:   c.ProjectName,
					SheetName:     sheetName[1:],
					Data:          &data,
					OptionalFiles: c.OptionalFiles,
				}}
				m[fileName] = append(m[fileName], ct)
			}
		}
	}

	type GenerateData struct {
		FileName string
		Sheets   []sheetType
	}
	g := []GenerateData{}

	for key := range m {
		sort.Slice(m[key], func(i, j int) bool {
			_, isEnumI := m[key][i].(*enumType)
			_, isEnumJ := m[key][j].(*enumType)

			if isEnumI != isEnumJ {
				return isEnumI && !isEnumJ
			}

			strI := fmt.Sprintf("%v", m[key][i])
			strJ := fmt.Sprintf("%v", m[key][j])

			return strI < strJ
		})

		g = append(g, GenerateData{
			FileName: key,
			Sheets:   m[key],
		})
	}

	errs := []error{}
	docs := map[string]string{}
	for _, d := range g {
		var preContent string
		include := map[string]any{}
		forwardContent := map[string]any{}
		var content string

		preContent += "// 이 파일은 자동으로 생성된 파일입니다. 수동으로 수정하지 마세요.\n"
		preContent += "\n"
		preContent += "#pragma once\n"
		preContent += "\n"
		preContent += "#include \"CoreMinimal.h\""

		sheetErrs := []error{}
		for _, sheet := range d.Sheets {
			c, p, i, err := sheet.Generate()
			if err != nil {
				sheetErrs = append(sheetErrs, err)
				break
			}
			content += c
			for _, v := range i {
				include[v] = nil
			}
			for _, v := range p {
				forwardContent[v] = nil
			}
		}
		if len(sheetErrs) > 0 {
			errs = append(errs, sheetErrs...)
			break
		}

		var result string
		result += preContent
		result += "\n"
		for key := range include {
			result += fmt.Sprintf("#include \"%s\"\n", key)
		}

		result += "\n"
		result += fmt.Sprintf("#include \"%s.generated.h\"\n", d.FileName)
		result += "\n"

		forwardContentArr := []string{}
		for key := range forwardContent {
			forwardContentArr = append(forwardContentArr, key)
		}

		sort.Strings(forwardContentArr)
		for _, v := range forwardContentArr {
			result += v + ";\n"
		}

		result += "\n"
		result += content

		docs[d.FileName] = result

	}

	var r error
	for _, err := range errs {
		r = errors.Join(r, err)
	}

	if r == nil {
		for fileName, doc := range docs {
			path := filepath.Join(c.DestDir, fileName+".h")
			h, err := os.Create(path)
			if err != nil {
				r = errors.Join(r, err)
				continue
			}

			if _, err := h.WriteString(doc); err != nil {
				errs = append(errs, err)
				h.Close()
				continue
			}
			h.Close()
		}
	}

	// 추가 옵션
	for _, d := range g {
		for _, sheet := range d.Sheets {
			sheet.Optional()
		}
	}

	return r
}
