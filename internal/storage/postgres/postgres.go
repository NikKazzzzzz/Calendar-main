package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/NikKazzzzzz/Calendar-main/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"time"
)

type Storage struct {
	db *sql.DB
}

func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &Storage{db: db}
	if err := storage.initDB(dbPath); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) initDB(dbPath string) error {
	migrationsDir := "./migrations"
	if err := applyMigrations(dbPath, migrationsDir); err != nil {
		return err
	}
	return nil
}

func (s *Storage) AddEvent(event *models.Event) error {
	query := `INSERT INTO events (title, description, start_time, end_time) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, event.Title, event.Description, event.StartTime.Format(time.RFC3339), event.EndTime.Format(time.RFC3339))
	return err
}

func (s *Storage) UpdateEvent(event *models.Event) error {
	query := `UPDATE events SET title = ?, description = ?, start_time = ?, end_time = ? WHERE id = ?`
	_, err := s.db.Exec(query, event.Title, event.Description, event.StartTime.Format(time.RFC3339), event.EndTime.Format(time.RFC3339), event.ID)
	return err
}

func (s *Storage) DeleteEvent(id int) error {
	query := `DELETE FROM events WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *Storage) GetEventByID(id int) (*models.Event, error) {
	event := &models.Event{}
	query := `SELECT id, title, description, start_time, end_time FROM events WHERE id = ?`
	row := s.db.QueryRow(query, id)
	var startTimeStr, endTimeStr string
	err := row.Scan(&event.ID, &event.Title, &event.Description, &startTimeStr, &endTimeStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	event.StartTime, err = time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return nil, err
	}

	event.EndTime, err = time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *Storage) IsTimeSlotTaken(startTime time.Time, endTime time.Time) (bool, error) {
	query := `SELECT COUNT(*) FROM events WHERE
            (start_time < ? AND end_time > ?) OR 
            (start_time >= ? AND end_time <= ?)`
	var count int
	err := s.db.QueryRow(query, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), startTime.Format(time.RFC3339), endTime.Format(time.RFC3339)).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func applyMigrations(dbPath string, migrationsDir string) error {
	// Создаем мигратор с указанием источника миграций и строки подключения к базе данных
	m, err := migrate.New(
		"file://"+migrationsDir,             // источник миграций
		fmt.Sprintf("sqlite3://%s", dbPath), // строка подключения к базе данных
	)
	if err != nil {
		return err
	}

	// Применяем миграции
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil // нет новых миграций
		}
		return err // ошибка применения миграций
	}

	return nil
}
