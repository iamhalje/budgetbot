package db

import (
	"database/sql"
	"time"
)

type DB struct {
	DB *sql.DB
}

type User struct {
	TelegramID    int64
	GitHubLogin   string
	MonthlyBudget float64
	Spent         float64
	BudgetMonth   string // "YYYY-MM"
}

func InitDB(db *sql.DB) error {
	sqlStatement := `
	CREATE TABLE IF NOT EXISTS users (
		telegram_id INTEGER PRIMARY KEY,
		github_login TEXT UNIQUE,
		monthly_budget REAL DEFAULT 0,
		spent REAL DEFAULT 0,
		budget_month TEXT
	)
	`
	_, err := db.Exec(sqlStatement)
	return err
}

// Добавлять/Обновлять пользователя по telegram_id
func UpdateUser(db *sql.DB, telegramID int64, githubLogin string) error {
	_, err := db.Exec(`
	INSERT INTO users (telegram_id, github_login, budget_month)
	VALUES (?, ?, ?)
	ON CONFLICT(telegram_id) DO UPDATE SET github_login=excluded.github_login
	`, telegramID, githubLogin, time.Now().Format("2006-01"))
	return err
}

func GetUserByTelegramID(db *sql.DB, telegramID int64) (*User, error) {
	row := db.QueryRow(`SELECT telegram_id, github_login, monthly_budget, spent, budget_month FROM users WHERE telegram_id = ?`, telegramID)
	u := User{}
	err := row.Scan(&u.TelegramID, &u.GitHubLogin, &u.MonthlyBudget, &u.Spent, &u.BudgetMonth)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Проверка, существует ли github_login в БД
func ExistsGithubLogin(db *sql.DB, githubLogin string) (bool, error) {
	row := db.QueryRow(`SELECT 1 FROM users WHERE github_login = ?`, githubLogin)
	var dummy int
	err := row.Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Устанавливаем/Обновляем бюджет на месяце, НЕ сбрасывая траты
func SetBudget(db *sql.DB, telegramID int64, budget float64, month string) error {
	_, err := db.Exec(`UPDATE users SET monthly_budget = ?, budget_month = ? WHERE telegram_id = ?`, budget, month, telegramID)
	return err
}

// Сбрасываем траты, если начался новый месяц
func ResetIfNewMonth(db *sql.DB, u *User) (bool, error) {
	currentMonth := time.Now().Format("2006-01")
	if u.BudgetMonth != currentMonth {
		_, err := db.Exec(`
			UPDATE users SET spent = 0, budget_month = ? WHERE telegram_id = ?
		`, currentMonth, u.TelegramID)
		if err != nil {
			return false, err
		}
		// Обновляем структуру пользователя в памяти
		u.Spent = 0
		u.BudgetMonth = currentMonth
		return true, nil
	}
	return false, nil
}

// Сбросить расходы вручную
func ResetSpent(db *sql.DB, telegramID int64) error {
	_, err := db.Exec(`UPDATE users SET spent = 0 WHERE telegram_id = ?`, telegramID)
	return err
}

// Обновить траты
func UpdateSpent(db *sql.DB, userID int64, newSpent float64) error {
	_, err := db.Exec("UPDATE users SET spent = ? WHERE telegram_id = ?", newSpent, userID)
	return err
}
