package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheet[T any] []T

func (l *Sheet[T]) UnmarshalText(b []byte) error {
	srv, err := sheets.NewService(context.Background(),
		option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(os.Getenv("SHEET_CREDENTIALS"))),
	)
	if err != nil {
		return err
	}

	url := string(b)
	sid, tab, err := parseSheetURL(srv, url)
	if err != nil {
		return err
	}

	result, err := getValues[T](srv, sid, tab)
	if err != nil {
		return err
	}

	*l = result
	return nil
}

var (
	reSpreadsheetURL = regexp.MustCompile(`/spreadsheets/d/([a-zA-Z0-9_-]+)/.+?gid=(\d+)`)
)

func parseSheetURL(srv *sheets.Service, url string) (sid string, tab string, err error) {
	m := reSpreadsheetURL.FindStringSubmatch(url)
	if len(m) == 0 {
		return "", "", fmt.Errorf("invalid sheet URL: %s", url)
	}
	sid = m[1]
	gid, _ := strconv.ParseInt(m[2], 10, 64)

	ss, err := srv.Spreadsheets.Get(sid).Fields("sheets.properties").Do()
	if err != nil {
		return "", "", fmt.Errorf("spreadsheets.get: %w", err)
	}
	for _, sheet := range ss.Sheets {
		if sheet.Properties.SheetId == gid {
			tab = sheet.Properties.Title
			break
		}
	}

	return sid, tab, nil
}

func getValues[T any](srv *sheets.Service, sid, tab string) ([]T, error) {
	if tab != "" {
		tab = fmt.Sprintf("'%s'!", tab)
	}

	res, err := srv.Spreadsheets.Values.BatchGet(sid).Ranges(tab + "A1:E6448").Do()
	if err != nil {
		return nil, fmt.Errorf("spreadsheets.batchGet(): %w", err)
	}

	typ := reflect.TypeFor[T]()

	var result []T
	var fields []string
	for y, col := range res.ValueRanges[0].Values {
		if y == 0 {
			for _, val := range col {
				if val == "" {
					break
				}
				fields = append(fields, val.(string))
			}
			continue
		}

		item := reflect.New(typ).Elem()
		for x, val := range col {
			if x >= len(fields) {
				break
			}

			f := item.FieldByName(fields[x])
			if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
				if strVal, ok := val.(string); ok {
					f.SetString(strVal)
				}
			}
		}

		result = append(result, item.Interface().(T))
	}

	return result, nil
}
