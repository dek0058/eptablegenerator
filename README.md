# eptablegenerator&nbsp;[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) ![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=flat&logo=go&logoColor=white) ![Go Version](https://img.shields.io/badge/Version-1.25.4-00ADD8?style=flat&logo=go&logoColor=white)

이 프로젝트는 데이터 파일을 게임 프로젝트의 소스 코드로 변환하여 개발자의 편의를 향상시키기 위해 생성되었습니다. 현재는 XLSX 파일을 Unreal Engine 5에서 사용할 수 있도록 `.h` 파일로 변환하여 저장합니다.

```markdown
# package table
├ config
└ xlsx

데이터 파일을 변환하기 위한 기본 자료 구조를 구현합니다.
```

### Config 구조

```markdown
project_name: (옵션)
source_dir: 변환할 데이터 파일 디렉토리 경로 (없을 시 실행 경로)
dest_dir: 생성할 데이터 파일 디렉토리 경로 (없을 시 실행 경로)
```

## 📄 package ueproject

XLSX 파일을 Unreal Engine에서 사용할 수 있도록 구조체 및 열거형 자료구조를 생성합니다.

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
