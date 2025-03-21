package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		*d = Date{v.UTC().Truncate(24 * time.Hour)}
		return nil
	case nil:
		*d = Date{}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Date", v)
	}
}

func (d Date) Value() (driver.Value, error) {
	return d.Time.UTC().Truncate(24 * time.Hour), nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Time.UTC().Format(time.DateOnly))), nil
}

func (d *Date) UnmarshalJSON(bs []byte) error {
	var s string
	err := json.Unmarshal(bs, &s)
	if err != nil {
		return err
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return err
	}
	*d = Date{t.UTC()}
	return nil
}

type Todo struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	DueTime     Date     `json:"due_time"`
}
