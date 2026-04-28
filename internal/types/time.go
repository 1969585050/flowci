package types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type JSONTime struct {
	time.Time
}

func NowJSON() JSONTime {
	return JSONTime{Time: time.Now().UTC()}
}

func NewJSONTime(t time.Time) JSONTime {
	return JSONTime{Time: t}
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Time.Format(time.RFC3339) + `"`), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		t.Time = time.Time{}
		return nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("parse JSONTime: %w", err)
	}
	t.Time = parsed
	return nil
}

func (t *JSONTime) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case string:
		parsed, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, v)
			if err != nil {
				return fmt.Errorf("scan JSONTime: %w", err)
			}
		}
		t.Time = parsed
		return nil
	}
	return fmt.Errorf("cannot scan %T into JSONTime", value)
}

func (t JSONTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}
