package handler

import (
	"database/sql"
	"fmt"
	"log"

	"bpeecs.nchu.edu.tw/config"
	_ "github.com/mattn/go-sqlite3"
)

type Calendar struct {
	ID    int64  `json:"id"`
	Year  uint   `json:"year"`
	Month uint   `json:"month"`
	Day   uint   `json:"day"`
	Event string `json:"event"`
	Link  string `json:"link"`
}

func (c *Calendar) Add() error {
	d, err := sql.Open("sqlite3", config.CalendarDB)
	if err != nil {
		return fmt.Errorf("sql.Open() error %v", err)
	}
	defer d.Close()
	stmt, err := d.Prepare("INSERT INTO calendar(year, month, day, event, link) values(?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("d.Prepare() error %v", err)
	}

	res, err := stmt.Exec(c.Year, c.Month, c.Day, c.Event, c.Link)
	if err != nil {
		return fmt.Errorf("stmt.Exec() error %v", err)
	}
	c.ID, err = res.LastInsertId()
	if err != nil {
		return fmt.Errorf("res.LastInsertId() error %v", err)
	}
	return nil
}

func (c *Calendar) Del() error {
	d, err := sql.Open("sqlite3", config.CalendarDB)
	if err != nil {
		return fmt.Errorf("sql.Open() error %v", err)
	}
	defer d.Close()
	stmt, err := d.Prepare("DELETE FROM calendar WHERE ID = ?")
	if err != nil {
		return fmt.Errorf("d.Prepare() error %v", err)
	}
	_, err = stmt.Exec(c.ID)
	if err != nil {
		return fmt.Errorf("stmt.Exec() error %v", err)
	}
	return nil
}

func (c *Calendar) Update() error {
	d, err := sql.Open("sqlite3", config.CalendarDB)
	if err != nil {
		return fmt.Errorf("sql.Open() error %v", err)
	}
	defer d.Close()
	stmt, err := d.Prepare(`UPDATE calendar SET year = ?, month = ?, day = ?, 
	event = ?, link = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("d.Prepare() error %v", err)
	}
	_, err = stmt.Exec(c.Year, c.Month, c.Day, c.Event, c.Link, c.ID)
	if err != nil {
		return fmt.Errorf("stmt.Exec() error %v", err)
	}
	return nil
}

func GetLatestCalendar(from int, to int) (list []Calendar, hasNext bool) {
	d, err := sql.Open("sqlite3", config.CalendarDB)
	if err != nil {
		log.Printf("sql.Open() error %v\n", err)
		return list, false
	}
	rows, err := d.Query(`SELECT id, year, month, day, event, link FROM calendar 
	ORDER BY YEAR DESC, MONTH DESC, DAY DESC
	LIMIT ?, ?`, from, to-from+2)
	if err != nil {
		log.Println(err)
		return nil, false
	}

	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		var info Calendar
		rows.Scan(&info.ID, &info.Year, &info.Month, &info.Day, &info.Event, &info.Link)

		if i == to-from+1 {
			hasNext = true
		} else {
			list = append(list, info)
		}
	}
	return list, hasNext
}

func GetCalendarByYearMonth(year uint, month uint) (list []Calendar) {
	d, err := sql.Open("sqlite3", config.CalendarDB)
	if err != nil {
		log.Printf("sql.Open() error %v\n", err)
		return list
	}

	rows, err := d.Query(`SELECT id, day, event, link FROM calendar WHERE year = ? and month = ?
	ORDER BY day ASC`, year, month)

	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		info := Calendar{
			Year:  year,
			Month: month,
		}
		rows.Scan(&info.ID, &info.Day, &info.Event, &info.Link)
		list = append(list, info)
	}
	return list
}
