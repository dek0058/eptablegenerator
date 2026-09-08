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

// TestGeneratorMSWStripsMakerBOM 은 메이커가 다시 저장한 파일을 원래 형태로 되돌리는지,
// 그리고 이미 같은 내용이면 다시 쓰지 않는지 확인합니다.
//
// MSW 메이커는 데이터셋을 열거나 새로고침할 때 .csv 를 UTF-8 BOM + CRLF 로 다시 저장합니다.
// 표의 내용은 그대로이므로, 생성기를 다시 돌리면 BOM 이 없는 원래 형태로 돌아와야 합니다.
func TestGeneratorMSWStripsMakerBOM(t *testing.T) {
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

	// 생성 직후의 내용을 기억해 둡니다.
	want := map[string][]byte{}
	for _, file := range csvFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		want[file] = content

		if bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}) {
			t.Fatalf("%s was generated with a BOM", filepath.Base(file))
		}
	}

	// 메이커가 다시 저장한 상태를 흉내냅니다 (BOM + CRLF).
	for _, file := range csvFiles {
		rewritten := append([]byte{0xEF, 0xBB, 0xBF}, bytes.ReplaceAll(want[file], []byte("\n"), []byte("\r\n"))...)
		if err := os.WriteFile(file, rewritten, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 다시 생성하면 BOM 과 개행이 원래대로 돌아와야 합니다.
	if err := gen.Generate(c); err != nil {
		t.Fatal(err)
	}

	for file, expected := range want {
		got, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, expected) {
			t.Errorf("%s was not restored to the generated form", filepath.Base(file))
		}
	}

	// 이미 같은 내용이면 파일을 건드리지 않아야 합니다.
	stats := map[string]os.FileInfo{}
	for _, file := range csvFiles {
		info, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}

		stats[file] = info
	}

	if err := gen.Generate(c); err != nil {
		t.Fatal(err)
	}

	for file, before := range stats {
		after, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}

		if !after.ModTime().Equal(before.ModTime()) {
			t.Errorf("%s was rewritten even though its contents did not change", filepath.Base(file))
		}
	}
}
