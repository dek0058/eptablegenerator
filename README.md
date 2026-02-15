# eptablegenerator&nbsp;[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) ![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=flat&logo=go&logoColor=white) ![Go Version](https://img.shields.io/badge/Version-1.25.4-00ADD8?style=flat&logo=go&logoColor=white)

이 프로젝트는 데이터 파일을 게임 프로젝트의 소스 코드로 변환하여 개발자의 편의를 향상시키기 위해 생성되었습니다. 현재는 XLSX 파일을 Unreal Engine 5에서 사용할 수 있도록 `.h` 파일로 변환하여 저장합니다.

## 목차

- [📄 package ueproject](#-package-ueproject)
- [📄 package msw](#-package-msw)

## 📄 package ueproject

XLSX 파일을 Unreal Engine에서 사용할 수 있도록 구조체 및 열거형 자료구조를 생성합니다.

### Config
```yaml
project_name: UnrealProject 프로젝트 모듈 이름 입니다 (ex: MYPROJECT)
source_dir: xlsx 파일이 위치한 디렉토리 입니다 (ex: ./data)
dest_dir: 생성된 .h 파일이 저장될 디렉토리 입니다 (ex: ./Generated)
```

### 예제

```markdown
TestStructTable.xlsx
├ !TestStruct
├ @TestEnum
└ #TestConst
```

#### !TestText
| Index | Name  | Value1 | Value2 |
|-------|-------|--------|--------|
| int32 | FText | int32  | float64|
| 1     | A     | 10     | 1.0    |
| 2     | B     | 20     | 2.0    |
| 3     | C     | 30     | 3.0    |
| 4     | D     | 40     | 4.0    |

```cpp
USTRUCT(BlueprintType)
struct FTestText : public FTableRowBase
{
    GENERATED_BODY()

    UPROPERTY(EditAnywhere, BlueprintReadWrite)
    int32 Index = INDEX_NONE;

    UPROPERTY(EditAnywhere, BlueprintReadWrite)
    FText Name;

    UPROPERTY(EditAnywhere, BlueprintReadWrite)
    int32 Value1 = INDEX_NONE;

    UPROPERTY(EditAnywhere, BlueprintReadWrite)
    double Value2 = INDEX_NONE;
};
```

#### @TestEnum
| Name  | Value | Memo |
|-------|-------|------|
| int32 | FString |      |
| 1     | EnumA  | A    |
| 2     | EnumB  | B    |
| 3     | EnumC  | C    |
| 4     | EnumD  | D    |

#### Output

```cpp
UENUM(BlueprintType)
enum class ETestEnum : uint8
{
    EnumA = 1 UMETA(DisplayName = "A"),
    EnumB = 2 UMETA(DisplayName = "B"),
    EnumC = 3 UMETA(DisplayName = "C"),
    EnumD = 4 UMETA(DisplayName = "D"),
    Max UMETA(Hidden)
};
ENUM_RANGE_BY_COUNT(ETestEnum, ETestEnum::Max)
```
#### #TestConst
| Type    | Name        | Value  |
|---------|-------------|--------|
| FString | DefaultName | Steven |
| int32   | Hp          | 200    |

#### Output

```cpp
UCLASS(config = Game, defaultconfig)
class TEST_API UConst : public UDeveloperSettings
{
    GENERATED_BODY()

public:
    UPROPERTY(Config, VisibleDefaultsOnly, BlueprintReadOnly, Category = "Table")
    FString DefaultName;

    UPROPERTY(Config, VisibleDefaultsOnly, BlueprintReadOnly, Category = "Table")
    int32 Hp;
};
```

## 📄 package msw

XLSX 파일을 MapleStoryWorlds에서 사용 할 수 있도록 구조체 및 CSV 파일을 생성합니다.

 - XLSX의 첫 번째 행은 데이터의 이름 입니다.
 - XLSX의 두 번째 행은 데이터의 타입 입니다.
 - XLSX의 세 번째 행은 데이터의 속성 입니다. 키값으로 쓰일지 컬럼 설명으로 쓰일지를 결정 합니다.
   - key: 해당 컬럼이 레코드의 키값으로 쓰입니다. 테이블에서 레코드를 가져올 때 사용됩니다.
   - design: 해당 컬럼은 레코드의 설명으로 쓰입니다. 테이블에서 레코드를 가져올 때 사용되지 않습니다.

### Config
```yaml
msw_config: WorldConfig.config 경로 입니다. root/Global에 위치해 있습니다.
source_dir: xlsx 파일이 위치한 디렉토리 입니다 (ex: ./data)
dest_dir: 생성된 .mlua 파일이 저장될 디렉토리 입니다 (ex: ./Generated)
csv_dir: 생성된 .csv 파일이 저장될 디렉토리 입니다 (ex: ./Generated/CSV)
```

### 예제

```markdown
TestStructTable.xlsx
└ Item
```

### Input
#### Item.xlsx (Item Sheet)
|Index    | Category  | Name | ItemDesc | Level  | SellGold | Equip  |
|---------|-----------|------|--------|----------|--------|--------|
| string  | string    |string| string | integer| number   | boolean|
| key | all | all | design | all | all | all |
| item_ironsword | Weaepon | 철검 | 철로 된 검 입니다 | 10 | 50 | TRUE |
| item_woodenarmor | Armor | 나무갑옷 | 나무로 된 갑옷 입니다 | 10 | 19.99 | FALSE |
| item_book | Miscellaneous | 책 | 책 입니다 | 1 | 3.33 | FALSE |

### Output

#### ItemRecord
```lua
---@description "자동 생성된 테이블 입니다. 수정하지 마세요 (Derivative: Item)"
@Struct
script ItemRecord

	property string Index = ""
	property string Category = ""
	property string Name = ""
	property integer Level = 0
	property number SellGold = 0.0
	property boolean Equip = false

end
```

#### ItemTable
```lua
---@description "자동 생성된 테이블 입니다. 수정하지 마세요 (Derivative: Item)"
@Struct
script ItemTable

	property table records = {}

	---@description "테이블을 로드 합니다"
	method void Load()
		local userDataset = _DataService:GetTable("ItemTable")
		local rowCount = userDataset:GetRowCount()

		for row = 1, rowCount do
			local record = ItemRecord()
			local cell = ""
			record.Index = userDataset:GetCell(row, "Index")
			record.Category = userDataset:GetCell(row, "Category")
			record.Name = userDataset:GetCell(row, "Name")
			cell = userDataset:GetCell(row, "Level")
			record.Level = tonumber(cell) or 0
			cell = userDataset:GetCell(row, "SellGold")
			record.SellGold = tonumber(cell) or 0.0
			cell = userDataset:GetCell(row, "Equip")
			record.Equip = (string.lower(cell) == "true")
			self.records[record.Index] = record
		end
	end

	---@description "테이블에서 레코드를 가져옵니다"
	method ItemRecord GetRecord(string key)
		return self.records[key]
	end

end

```

#### ItemTable.csv
```
Index,Category,Name,Level,SellGold,Equip
item_ironsword,Weaepon,철검,10,50,TRUE
item_woodenarmor,Armor,나무갑옷,10,19.99,TRUE
item_book,Miscellaneous,책,1,3.33,FALSE
```
