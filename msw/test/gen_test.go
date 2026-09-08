package test

import (
	"bytes"
	"eptablegenerator/msw/config"
	"eptablegenerator/msw/gen"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestGeneratorMSW(t *testing.T) {
	t.Log("TestGeneratorMSW")

	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	c := &config.Config{
		MswConfig: path.Join(p, "WorldConfig.config"),
		SourceDir: p,
		DestDir:   p,
		CsvDir:    p,
	}

	if err := gen.Generate(c); err != nil {
		t.Fatal(err)
	}
}

// TestGeneratorMSWKeepsMakerBOM 은 내용이 바뀌지 않은 파일을 다시 쓰지 않는지 확인합니다.
//
// MSW 메이커는 데이터셋을 열거나 새로고침할 때 .csv 를 UTF-8 BOM + CRLF 로 다시 저장합니다.
// 이때 생성기가 파일을 무조건 덮어쓰면 표의 내용이 그대로인데도 형상 관리에서는
// 매번 변경된 파일로 잡히게 됩니다.
func TestGeneratorMSWKeepsMakerBOM(t *testing.T) {
	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	out := t.TempDir()

	c := &config.Config{
		MswConfig: path.Join(p, "WorldConfig.config"),
		SourceDir: p,
		DestDir:   out,
		CsvDir:    out,
	}

	if err := gen.Generate(c); err != nil {
		t.Fatal(err)
	}

	csvFiles, err := filepath.Glob(filepath.Join(out, "*.csv"))
	if err != nil {
		t.Fatal(err)
	}

	if len(csvFiles) == 0 {
		t.Fatal("no csv files were generated")
	}

	// 메이커가 다시 저장한 상태를 흉내냅니다 (BOM + CRLF).
	maker := map[string][]byte{}
	for _, file := range csvFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		rewritten := append([]byte{0xEF, 0xBB, 0xBF}, bytes.ReplaceAll(content, []byte("\n"), []byte("\r\n"))...)
		if err := os.WriteFile(file, rewritten, 0644); err != nil {
			t.Fatal(err)
		}

		maker[file] = rewritten
	}

	if err := gen.Generate(c); err != nil {
		t.Fatal(err)
	}

	for file, want := range maker {
		got, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, want) {
			t.Errorf("%s was rewritten even though its contents did not change", filepath.Base(file))
		}
	}
}
