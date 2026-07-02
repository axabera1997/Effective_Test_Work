package models

import (
	"encoding/json"
	"errors"
	"time"
)

// переопределение метода анмаршаллинга для типа Subscription
func (sub *Subscription) UnmarshalJSON(data []byte) (err error) {

	// Создаем временный тип, чтобы избежать рекурсии при Unmarshal
	type Alias Subscription
	temp := &struct {
		// StartDate string `json:"start_date"`
		// EndDate   string `json:"end_date"`
		*Alias
	}{
		Alias: (*Alias)(sub),
	}
	// Разбираем JSON во временную структуру
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	// дата из строки - в time.Time
	sub.Sdt, err = ParseDate(temp.Start_date)
	// sub.Sdt, err = ParseDate(temp.StartDate)
	if err != nil {
		return err
	}
	sub.Edt, err = ParseDate(temp.End_date)
	// sub.Edt, err = ParseDate(temp.EndDate)
	if err != nil {
		return err
	}

	return
}

// ParseDate принимает строковую дату, возвращает time.Time или nil. Отсекает день месяца, устанавливает в 1е число месяца - подписка помесячно, даты не важны
func ParseDate(date string) (time.Time, error) {
	// если дата пустая
	if date == "" {
		return time.Time{}, nil
	}

	// парсим месяц-год с двухзначным годом (например, 01-25)
	if t, err := time.Parse("01-06", date); err == nil {
		year := t.Year()
		if year < 100 {
			year += 2000 // интерпретируем как 2000+год
		}
		return time.Date(year, t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}

	// парсим месяц-год с четырьмя цифрами (например, 01-2025)
	if t, err := time.Parse("01-2006", date); err == nil {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}

	// парсим по стандарту RFC3339, игнорируем время и день
	if t, err := time.Parse(time.RFC3339, date); err == nil {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, errors.New("неверный формат даты")
}
