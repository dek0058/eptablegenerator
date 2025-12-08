package xlsx

import (
	"errors"

	"github.com/xuri/excelize/v2"
)

func NewXLSX(filepath string) *XLSX {
	m := &XLSX{
		Data: make(map[string][][]string),
	}

	if len(filepath) > 0 {
		if err := m.Load(filepath); err != nil {
			panic(err)
		}
	}

	return m
}

type XLSX struct {
	Data map[string][][]string
}

func (m *XLSX) Load(filePath string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	errs := []error{}
	for _, sheet := range f.WorkBook.Sheets.Sheet {
		if err := m.addSheet(f, sheet.Name); err != nil {
			errs = append(errs, err)
			continue
		}
	}

	var r error
	for _, err := range errs {
		r = errors.Join(r, err)
	}
	return r
}

func (m *XLSX) addSheet(f *excelize.File, sheetName string) error {
	if f == nil {
		return errors.New("file is nil")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}

	m.Data[sheetName] = rows

	return nil
}
