package postgres

import (
	"database/sql"
	"errors"
	"github.com/NikKazzzzzz/Calendar-main/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"log/slog"
	"time"
)

type Storage struct {
	db  *sql.DB
	log *slog.Logger
}

func New(dsn string, log *slog.Logger) (*Storage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	storage := &Storage{db: db, log: log}
	if err := storage.initDB(dsn); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) initDB(dsn string) error {
	migrationsDir := "./migrations"
	s.log.Debug("Applying migrations from directory:", slog.String("dir", migrationsDir))
	if err := applyMigrations(dsn, migrationsDir); err != nil {
		return err
	}
	return nil
}

func (s *Storage) AddEvent(event *models.Event) error {
	query := `INSERT INTO events (title, description, start_time, end_time) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, event.Title, event.Description, event.StartTime, event.EndTime)
	return err
}

func (s *Storage) UpdateEvent(event *models.Event) error {
	query := `UPDATE events SET title = $1, description = $2, start_time = $3, end_time = $4 WHERE id = $5`
	_, err := s.db.Exec(query, event.Title, event.Description, event.StartTime, event.EndTime, event.ID)
	return err
}

func (s *Storage) DeleteEvent(id int) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *Storage) GetEventByID(id int) (*models.Event, error) {
	event := &models.Event{}
	query := `SELECT id, title, description, start_time, end_time FROM events WHERE id = $1`
	row := s.db.QueryRow(query, id)
	var startTimeStr, endTimeStr string
	err := row.Scan(&event.ID, &event.Title, &event.Description, &startTimeStr, &endTimeStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return event, nil
}

func (s *Storage) IsTimeSlotTaken(startTime time.Time, endTime time.Time) (bool, error) {
	query := `SELECT COUNT(*) FROM events WHERE
            (start_time < $1 AND end_time > $2) OR 
            (start_time >= $3 AND end_time <= $4)`
	var count int
	err := s.db.QueryRow(query, startTime, endTime, startTime, endTime).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func applyMigrations(dsn string, migrationsDir string) error {
	// Создаем мигратор с указанием источника миграций и строки подключения к базе данных
	m, err := migrate.New(
		"file://"+migrationsDir, // источник миграций //
		dsn,                     // строка подключения к базе данных
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
